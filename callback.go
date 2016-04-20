package pilosa

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

var MAX_BATCH_SIZE = 200

func callbackResults(req *QueryRequest, resultsChan chan CallRes) {
	resBatch := make([]interface{}, 0)
	for {
		select {
		case res, more := <-resultsChan:
			if !more {
				if len(resBatch) > 0 {
					callCallback(resBatch, req)
				}
				log.Printf("Completed: %v", req)
				return
			}
			resBatch = append(resBatch, res)
		case <-time.After(time.Second * 2):
			if len(resBatch) > 0 {
				callCallback(resBatch, req)
				// TODO return and check for err to make sure
				// callback endpoint is still alive only clear resBatch if callback
				// successful - otherwise, buffer results up to a point, and if callback
				// is still dead, give up
				resBatch = make([]interface{}, 0)
			}
		}
		if len(resBatch) > MAX_BATCH_SIZE {
			callCallback(resBatch, req)
			// TODO return and check for err to make sure
			// callback endpoint is still alive only clear resBatch if callback
			// successful - otherwise, buffer results up to a point, and if callback
			// is still dead, give up
			resBatch = make([]interface{}, 0)
		}
	}

}

func callCallback(resultBatch []interface{}, req *QueryRequest) {
	// TODO support protobuf as well as JSON
	b := bytes.Buffer{}
	err := json.NewEncoder(&b).Encode(resultBatch)
	if err != nil {
		log.Printf("Error: %v, couldn't marshal resultBatch: %v", err, resultBatch)
	}
	log.Printf("POSTING encoded = %s", string(b.Bytes()))
	resp, err := http.Post(req.CallbackURL, "application/json; charset=utf-8", &b)
	if err != nil {
		log.Printf("Error: %v, Couldn't post to callbackURL: %v", err, req.CallbackURL)
	}
	log.Println("RESP:! ", resp)
}
