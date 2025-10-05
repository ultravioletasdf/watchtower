package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"videoapp/internal/utils"

	"github.com/minio/minio-go/v7"
	gcache "github.com/patrickmn/go-cache"
	"github.com/rs/cors"
)

var s3 *minio.Client
var cache *gcache.Cache
var pubKey *ecdsa.PublicKey

func main() {
	cfg := utils.ParseConfig()

	readPub()

	cache = gcache.New(time.Hour*6, 10*time.Minute)

	s3 = utils.ConnectS3(cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{bucket}/{object...}", getResource)

	handler := cors.AllowAll().Handler(mux)
	http.ListenAndServe(":4000", handler)
}

type PresignedToken struct {
	Prefix   string
	ExpireAt time.Time
}

func readPub() {
	key, err := os.ReadFile("../../public_key.pem")
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PUBLIC KEY" {
		panic("Failed to decode public key")
	}
	pubIfc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	pub, ok := pubIfc.(*ecdsa.PublicKey)
	if !ok {
		panic("bad pub key")
	}
	pubKey = pub
}
func getResource(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	object := r.PathValue("object")
	if bucket == "" || object == "" {
		w.WriteHeader(400)
		return
	}
	pl := r.Header.Get("wt-payload")
	sig := r.Header.Get("wt-sig")

	if pl == "" || sig == "" {
		w.WriteHeader(400)
		return
	}

	var payload PresignedToken
	pl_json, err := base64.RawURLEncoding.DecodeString(pl)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}
	err = json.Unmarshal(pl_json, &payload)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}

	signature, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}

	hash := sha256.Sum256(pl_json)
	valid := ecdsa.VerifyASN1(pubKey, hash[:], signature)
	if !valid {
		w.WriteHeader(401)
		return
	}
	path := fmt.Sprintf("/%s/%s", bucket, object)
	if !strings.HasPrefix(path, payload.Prefix) {
		w.WriteHeader(403)
		return
	}
	if time.Now().After(payload.ExpireAt) {
		w.WriteHeader(403)
	}
	getObject(w, r, bucket, object)
}

func getObject(w http.ResponseWriter, r *http.Request, bucket, object string) {
	filename := "cache/" + url.QueryEscape(object)
	file, err := os.Open(filename)
	// File isn't in disk cache, fetch from S3
	if errors.Is(err, os.ErrNotExist) {
		obj, err := s3.GetObject(r.Context(), bucket, object, minio.GetObjectOptions{})
		if err != nil {
			w.WriteHeader(500)
			fmt.Println(err.Error())
			return
		}
		file, err := os.Create(filename)
		if err != nil {
			w.WriteHeader(500)
			fmt.Println(err.Error())
			return
		}

		w.Header().Set("Cache-Control", "max-age=31536000")

		// Copy to http writer and disk writer simultaniously
		multiwriter := io.MultiWriter(w, file)
		if _, err := io.Copy(multiwriter, obj); err != nil {
			w.WriteHeader(500)
			fmt.Println(err.Error())
			return
		}
	} else if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}

	w.Header().Set("Cache-Control", "max-age=31536000")
	// File was cached, serve
	if _, err := io.Copy(w, file); err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
		return
	}
}
