package main

import (
	"filestore-byceph/route"

	cfg "filestore-byceph/config"
)

func main() {
	router := route.Router()
	router.Run(cfg.UploadServiceHost)

}
