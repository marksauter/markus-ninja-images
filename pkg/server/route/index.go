package route

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/marksauter/markus-ninja-images/pkg/myconf"
	"github.com/marksauter/markus-ninja-images/pkg/myhttp"
	"github.com/marksauter/markus-ninja-images/pkg/mylog"
	"github.com/marksauter/markus-ninja-images/pkg/mytype"
	"github.com/marksauter/markus-ninja-images/pkg/service"
	"github.com/marksauter/markus-ninja-images/pkg/util"
	minio "github.com/minio/minio-go"
	"github.com/rs/cors"
)

type IndexHandler struct {
	Conf       *myconf.Config
	StorageSvc *service.StorageService
}

func (h IndexHandler) Cors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodOptions, http.MethodGet},
		AllowedOrigins: []string{h.Conf.ClientURL},
		Debug:          true,
	})
}

func (h IndexHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.Conf == nil || h.StorageSvc == nil {
		err := errors.New("route inproperly setup")
		mylog.Log.WithError(err).Error(util.Trace(""))
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	if req.Method != http.MethodGet {
		response := myhttp.MethodNotAllowedResponse(req.Method)
		myhttp.WriteResponseTo(rw, response)
		return
	}

	routeVars := mux.Vars(req)

	userID := routeVars["user_id"]
	uid, err := mytype.NewOIDFromShort("User", userID)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to create new oid")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	key := routeVars["key"]

	s := req.URL.Query().Get("s")
	var size int
	if s != "" {
		parsedSize, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to parse size")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
		size = int(parsedSize)
		if size != 400 {
			mylog.Log.WithError(err).Error("failed to parse size")
			response := myhttp.InvalidRequestErrorResponse("size may only be 400")
			myhttp.WriteResponseTo(rw, response)
			return
		}
	}

	var object *minio.Object
	defer object.Close()
	if s != "" {
		object, err = h.StorageSvc.GetThumbnail(size, uid, key)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to get thumbnail")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
	} else {
		object, err = h.StorageSvc.Get(uid, key)
		if err != nil {
			mylog.Log.WithError(err).Error("failed to get file")
			response := myhttp.InternalServerErrorResponse(err.Error())
			myhttp.WriteResponseTo(rw, response)
			return
		}
	}

	n, err := io.Copy(rw, object)
	if err != nil {
		mylog.Log.WithError(err).Error("failed to copy asset to response writer")
		response := myhttp.InternalServerErrorResponse(err.Error())
		myhttp.WriteResponseTo(rw, response)
		return
	}

	mylog.Log.WithField("size", n).Info("successfully streamed file")
	return
}
