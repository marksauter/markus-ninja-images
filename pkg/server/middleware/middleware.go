package middleware

import (
	"github.com/gorilla/handlers"
	"github.com/justinas/alice"
	"github.com/marksauter/markus-ninja-images/pkg/mylog"
)

var CommonMiddleware = alice.New(
	mylog.Log.AccessMiddleware,
	handlers.RecoveryHandler(),
)
