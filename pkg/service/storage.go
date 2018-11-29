package service

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	// Allow processing of images
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/marksauter/markus-ninja-images/pkg/myaws"
	"github.com/marksauter/markus-ninja-images/pkg/myconf"
	"github.com/marksauter/markus-ninja-images/pkg/mylog"
	"github.com/marksauter/markus-ninja-images/pkg/mytype"
	"github.com/marksauter/markus-ninja-images/pkg/util"
	minio "github.com/minio/minio-go"
	minioCreds "github.com/minio/minio-go/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// StorageService - service used for storing assets
type StorageService struct {
	bucket string
	svc    *minio.Client
}

// NewStorageService - create a new storage service instance
func NewStorageService(conf *myconf.Config) (*StorageService, error) {
	branch := util.GetRequiredEnv("BRANCH")
	var svc *minio.Client
	var err error
	if branch != "production" {
		credentials := minioCreds.NewFileAWSCredentials("", "")
		endPoint := "s3.amazonaws.com"
		useSSL := true
		svc, err = minio.NewWithCredentials(
			endPoint,
			credentials,
			useSSL,
			myaws.AWSRegion,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	} else {
		iam := minioCreds.NewIAM("")
		endPoint := "s3.amazonaws.com"
		useSSL := true
		svc, err = minio.NewWithCredentials(
			endPoint,
			iam,
			useSSL,
			myaws.AWSRegion,
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	// svc.TraceOn(nil)
	return &StorageService{
		bucket: conf.AWSUploadBucket,
		svc:    svc,
	}, nil
}

// Get - get an object from passed userID, and key
func (s *StorageService) Get(
	userID *mytype.OID,
	key string,
) (*minio.Object, error) {
	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		key[:2],
		key[3:5],
		key[6:8],
		key[9:],
	)
	objectPath := strings.Join(
		[]string{
			userID.Short,
			objectName,
		},
		"/",
	)

	mylog.Log.WithFields(logrus.Fields{
		"user_id": userID.Short,
		"key":     key,
	}).Info(util.Trace("object found"))
	return s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
}

// GetThumbnail - get a thumbnail of an asset, and generate it first if
// necessary
func (s *StorageService) GetThumbnail(
	size int,
	userID *mytype.OID,
	key string,
) (*minio.Object, error) {
	sizeStr := strconv.FormatInt(int64(size), 10)
	// Thumbnail objects are identified with a -'size' at the end of the key
	thumbKey := key + "--" + sizeStr
	thumbLocal := "./tmp/" + thumbKey + ".jpg"

	objectName := fmt.Sprintf(
		"%s/%s/%s/%s",
		thumbKey[:2],
		thumbKey[3:5],
		thumbKey[6:8],
		thumbKey[9:],
	)
	objectPath := strings.Join([]string{
		userID.Short,
		objectName,
	}, "/")

	objInfo, err := s.svc.StatObject(
		s.bucket,
		objectPath,
		minio.StatObjectOptions{},
	)
	if err != nil {
		minioError := minio.ToErrorResponse(err)
		if minioError.Code != "NoSuchKey" {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		mylog.Log.Info("generating new thumbnail...")

		asset, err := s.Get(userID, key)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		img, err := imaging.Decode(asset)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
		thumb := imaging.Thumbnail(img, size, size, imaging.CatmullRom)

		// create a new blank image
		dst := imaging.New(size, size, color.NRGBA{0, 0, 0, 0})

		// paste thumbnails into the new image
		dst = imaging.Paste(dst, thumb, image.Pt(0, 0))

		// ensure path is available
		dir := filepath.Dir(thumbLocal)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		err = imaging.Save(dst, thumbLocal)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		thumbFile, err := os.Open(thumbLocal)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		thumbStat, err := thumbFile.Stat()
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}

		_, err = s.svc.PutObject(
			s.bucket,
			objectPath,
			thumbFile,
			thumbStat.Size(),
			minio.PutObjectOptions{ContentType: objInfo.ContentType},
		)
		if err != nil {
			mylog.Log.WithError(err).Error(util.Trace(""))
			return nil, err
		}
	}

	mylog.Log.WithFields(logrus.Fields{
		"size":    size,
		"user_id": userID.Short,
		"key":     key,
	}).Info(util.Trace("thumbnail found"))
	return s.svc.GetObject(
		s.bucket,
		objectPath,
		minio.GetObjectOptions{},
	)
}
