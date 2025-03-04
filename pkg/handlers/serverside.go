package handlers

import (
	"bytes"
	"crypto/sha1"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/spencer-p/surfdash/pkg/data"
	"github.com/spencer-p/surfdash/pkg/meta"
	"github.com/spencer-p/surfdash/pkg/metrics"
	"github.com/spencer-p/surfdash/pkg/noaa"
	"github.com/spencer-p/surfdash/pkg/sunset"
	"github.com/spencer-p/surfdash/pkg/timetricks"
	"github.com/spencer-p/surfdash/pkg/visualize"
	"golang.org/x/crypto/pbkdf2"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	sessionName       = "good-times"
	sessionLastViewed = "last-viewed-referrer"
	minTideCookieName = "minTide"
	maxTideCookieName = "maxTide"
	userID            = "userid"
	// See https://developer.chrome.com/blog/cookie-max-age-expires.
	defaultMaxAge = 60 * 60 * 24 * 400 // 400 days in seconds.
)

var (
	store = &sessions.CookieStore{
		Codecs: securecookie.CodecsFromPairs(
			getSessionKey(),
			getEncryptionKey(),
			getSessionKey(),
			nil,
			[]byte("deadbeef"), // TODO: Remove.
			nil,
		),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   defaultMaxAge,
			Secure:   true,
			HttpOnly: true,
		},
	}
	db = data.PostgresFromEnvOrDie()
)

func init() {
	store.MaxAge(defaultMaxAge)
}

type TemplateInput struct {
	PresentationElements []PresentationElement
	NextStart            string
	PrevStart            string
	Name                 string
}

type PresentationElement struct {
	Date      string
	GoodTimes []meta.GoodTime
	TideImage template.HTML
}

// serverSideIndex serves a good times page fully rendered on the server.
func makeServerSideIndex(content embed.FS) http.HandlerFunc {
	indexTemplate := template.Must(template.ParseFS(content, "static/index.template.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)
		metrics.ObserveUserRequest(session.Values[userID])
		session.Values[sessionLastViewed] = r.URL.String()
		if err := session.Save(r, w); err != nil {
			log.Println("save session err", err)
		}

		date := time.Now()
		startString := r.FormValue("start")
		if startString != "" {
			parsed, err := time.Parse(time.RFC3339, startString)
			if err != nil {
				log.Printf("Failed to read time %q: %v", startString, err)
			} else {
				date = parsed
			}
		}

		// Fetch tide data first.
		query := noaa.PredictionQuery{
			// Add extra padding of one day around tides to fill in gaps.
			Start:    date.Add(-1 * 24 * time.Hour),
			Duration: forecastLength + 24*time.Hour,
			Station:  noaa.SantaCruz,
		}
		preds, err := noaa.GetPredictions(&query)
		if err != nil {
			err := fmt.Errorf("failed to fetch from NOAA: %w", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to fetch good times: %+v", err)
			log.Printf("Failed to fetch good times: %+v", err)
			return
		}

		// Compute sun events, goodtimes, and set up tide images.
		sunevents := sunset.GetSunEvents(date, query.Duration, sunset.SantaCruz)
		// Truncate the good times predictions to account for the
		// extra data data from above.
		trimIndex := lastIndexBefore(preds, timetricks.TrimClock(date.Add(forecastLength)))
		opts, _ := goodTimeOptionsFromSession(session)
		goodTimes := meta.GoodTimes2(meta.Conditions{preds[:trimIndex+1], sunevents}, opts)
		tideimages := visualize.NewTidal(preds, sunevents)

		presElems := goodTimesToPresentationElements(tideimages, goodTimes)

		tinput := TemplateInput{
			PresentationElements: presElems,
			NextStart:            date.Add(forecastLength).Format(time.RFC3339),
			PrevStart:            date.Add(-1 * forecastLength).Format(time.RFC3339),
		}

		w.Header().Add("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if err := indexTemplate.Execute(w, tinput); err != nil {
			log.Printf("Failed to execute template: %v", err)
		}
	})
}

func imgToString(img *visualize.Tidal, t time.Time) string {
	img.SetDate(t)
	var b bytes.Buffer
	img.Encode(&b)
	return b.String()
}

// TODO(spencer-p) Standardize these scattered binary search functions.
func lastIndexBefore(preds noaa.Predictions, t time.Time) int {
	left, right := 0, len(preds)
	for right-left > 1 {
		mid := (left + right) / 2
		midt := preds[mid].T()
		if midt.Before(t) {
			left = mid
		} else if midt.After(t) {
			right = mid
		} else if midt.Equal(t) {
			return mid
		}
	}
	ok := left < len(preds)
	if !ok {
		// Nothing found, just return last element
		return len(preds) - 1
	}
	return left
}

func goodTimesToPresentationElements(tideimages *visualize.Tidal, goodTimes []meta.GoodTime) []PresentationElement {
	var f func(result []PresentationElement, goodTimes []meta.GoodTime) []PresentationElement
	f = func(result []PresentationElement, goodTimes []meta.GoodTime) []PresentationElement {
		if len(goodTimes) == 0 {
			return result
		}

		resultLen := len(result)
		gt := goodTimes[0]
		gt.UpdatePrettyTime()

		if len(result) != 0 && result[resultLen-1].Date == timetricks.Day(gt.Time) {
			// There is already an entry in the result that corresponds to the
			// same day as the next time we're entering.
			result[resultLen-1].GoodTimes = append(result[resultLen-1].GoodTimes, gt)
		} else {
			// Normal case.
			result = append(result, PresentationElement{
				Date:      timetricks.Day(gt.Time),
				GoodTimes: []meta.GoodTime{gt},
				TideImage: template.HTML(imgToString(tideimages, gt.Time)),
			})
		}

		return f(result, goodTimes[1:])
	}

	return f(nil, goodTimes)
}

func goodTimeOptionsFromSession(s *sessions.Session) (meta.Options, *data.User) {
	opts := meta.Options{}

	id, ok := s.Values[userID]
	if !ok {
		return opts, nil
	}

	// Note the db lookup can fail here, and that's
	// fine. We'll just use default options.
	var user data.User
	if r := db.First(&user, id); r.Error != nil {
		log.Printf("Failed to find user %v: %v", id, r.Error)
	} else {
		// Log the time since we last saw the user.
		if !user.LastSeen.IsZero() {
			sinceLastUpdate := time.Since(user.LastSeen)
			log.Printf("User %d (%q) was last seen %s ago", id, user.Name, sinceLastUpdate)
		}
		user.LastSeen = time.Now()
		db.Save(&user)
	}
	opts.LowTideThresh = user.MinTide
	opts.HighTideThresh = user.MaxTide

	return opts, &user
}

func makeConfigTideParameters(redirectPrefix string, content embed.FS) http.HandlerFunc {
	configTideTemplate := template.Must(template.ParseFS(content, "static/config_tide.template.html"))

	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, sessionName)
		metrics.ObserveUserRequest(session.Values[userID])

		if r.Method == "GET" {
			session.Save(r, w)
			opts, user := goodTimeOptionsFromSession(session)
			opts.DefaultHighTide = ptr(float64(1))
			opts.DefaultLowTide = ptr(float64(-1000))
			if err := configTideTemplate.Execute(w, map[string]any{
				"Options": opts,
				"User":    user,
			}); err != nil {
				log.Printf("Failed to write configTideTemplate: %v", err)
			}
			return
		}
		// The remainder of this function assumes method is POST.
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse the form data.
		if err := r.ParseForm(); err != nil {
			msg := fmt.Sprintf("Failed to parse form: %v", err)
			log.Println(msg)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, msg)
			return
		}

		var user data.User
		if id, ok := session.Values[userID].(uint); ok {
			// Read-modify-write if the user provided an ID.
			// Otherwise, one will be generated with db.Save later.
			db.First(&user, id)
		}
		if f, err := strconv.ParseFloat(r.PostForm.Get("min_tide"), 64); err == nil {
			user.MinTide = &f
		} else {
			user.MinTide = nil
		}
		if f, err := strconv.ParseFloat(r.PostForm.Get("max_tide"), 64); err == nil {
			user.MaxTide = &f
		} else {
			user.MaxTide = nil
		}

		// Parse the birthday field.
		birthdayStr := r.PostForm.Get("birthday")
		if birthdayStr != "" {
			birthday, err := time.Parse("2006-01-02", birthdayStr)
			if err == nil {
				user.Birthday = birthday
			} else {
				log.Printf("Failed to parse birthday %q: %v", birthdayStr, err)
			}
		}

		// Log the time since the last update.
		if user.UpdatedAt.IsZero() {
			log.Printf("User %d (%q) has never been updated", id, user.Name)
		} else {
			sinceLastUpdate := time.Now().Sub(user.UpdatedAt)
			log.Printf("User %d (%q) was last updated %s ago", id, user.Name, sinceLastUpdate)
		}

		// Set the LastSeen column to the current time.
		user.LastSeen = time.Now()
		user.Name = r.PostForm.Get("name")
		if tx := db.Save(&user); tx.Error != nil {
			msg := fmt.Sprintf("Failed to save preferences: %v", tx.Error)
			log.Println(msg)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, msg)
			return
		}
		session.Values[userID] = user.ID
		session.Values["name"] = r.PostForm.Get("name")
		session.Save(r, w)

		// Redirect to whatever they saw last, or the index.
		referredFrom, ok := session.Values[sessionLastViewed].(string)
		if !ok || referredFrom == "/config" {
			referredFrom = "/"
		}
		redirectTo := pathJoinPreservePrefix(redirectPrefix, referredFrom)
		http.Redirect(w, r, redirectTo, http.StatusFound)
	}
}

func pathJoinPreservePrefix(prefix string, suffix string) string {
	trimmedPrefix := path.Join(prefix, "")
	result := path.Join(prefix, suffix)
	if result == trimmedPrefix {
		return prefix
	}
	return result
}

// getSessionKey returns a key to encrypt session cookies defined in the
// environment.
// If it is not set, it uses a compile-time default.
func getSessionKey() []byte {
	defaultKey := []byte("deadbeef")
	if key := os.Getenv("SESSION_KEY"); key != "" {
		return []byte(key)
	} else {
		return defaultKey
	}
}

func getEncryptionKey() []byte {
	password := "deadbeef"
	if fromEnv := os.Getenv("ENCRYPTION_KEY"); fromEnv != "" {
		password = fromEnv
	}
	return pbkdf2.Key([]byte(password), []byte{}, 4096, 32, sha1.New)
}

func ptr[T any](t T) *T {
	return &t
}
