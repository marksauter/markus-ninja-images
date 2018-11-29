package mylog

import (
	"net/http"
	"os"

	"github.com/marksauter/markus-ninja-images/pkg/util"
	"github.com/sirupsen/logrus"
)

var Log = New()

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	log := logrus.New()
	branch := util.GetRequiredEnv("BRANCH")
	forceColors := false
	if branch == "development.local" {
		forceColors = true
	}
	log.Formatter = &logrus.TextFormatter{ForceColors: forceColors}
	log.Out = os.Stdout
	if branch == "development.local" {
		log.SetLevel(logrus.DebugLevel)
	}
	return &Logger{log}
}

func (l *Logger) AccessMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		l.WithFields(logrus.Fields{
			"remote_addr": req.RemoteAddr,
			"method":      req.Method,
			"url":         req.URL,
			"proto":       req.Proto,
		}).Info("Request Info")
		l.WithField("user_agent", req.UserAgent()).Info("")
		// if l.Level >= logrus.DebugLevel {
		//   body, err := ioutil.ReadAll(req.Body)
		//   if err != nil {
		//     l.WithField("error", err).Error("Error reading request body")
		//   }
		//   reqStr := ioutil.NopCloser(bytes.NewBuffer(body))
		//   l.WithField("body", string(body)).Debug("")
		//   req.Body = reqStr
		// }
		h.ServeHTTP(rw, req)
	})
}
