package lookup

import (
        "log"
        "os/exec"
        "strings"
	"sync"
	"time"
)

var cache = make(map[string]cachedResult)
var cacheMutex sync.Mutex

type cachedResult struct {
	result string
	expires time.Time
}

const cacheDuration = 60 * time.Minute

func ExtLookupUser(user string, lookup string, l *log.Logger) (u string, err error) {

        // todo: this whole package needs some love, quite some love
        switch lookup {
        case "GECOS":
                u, err = lookupGECOS(user, l)
        case "username2slackID":
                u, err = lookupUsername2SlackID(user, l)
        default:
                u = user
        }

        return u, err
}

func lookupGECOS(u string, l *log.Logger) (string, error) {

        out, err := exec.Command("/usr/bin/getent", "passwd", u).Output()
        if err != nil {
                return u, err
        }
        fields := strings.Split(string(out), ":")

        return fields[4], nil
}

func lookupUsername2SlackID(u string, l *log.Logger) (string, error) {
	cacheMutex.Lock()
        defer cacheMutex.Unlock()

        // Check if the result is in the cache and not expired
        if cached, ok := cache[u]; ok && time.Now().Before(cached.expires) {
                // Return cached result
		l.Println("Found slackID in cache", cached.result)
                return cached.result, nil
        }
        out, err := exec.Command("/usr/local/bin/username2slackID", u).Output()
        if err != nil {
                return u, err
        }
	// The output is a byte slice, convert it to a string and trim spaces
        result := strings.TrimSpace(string(out))

        // Cache the result with expiration time
        cache[u] = cachedResult{result: result, expires: time.Now().Add(cacheDuration)}

        // Log the result if needed
        l.Println("Output of username2Slack:", result)

        return result, nil

}
