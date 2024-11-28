package gobotbsky

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	appbsky "github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"io"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
)

const defaultPDS = "https://bsky.social"

var blob []lexutil.LexBlob

// Wrapper over the atproto xrpc transport
type BskyAgent struct {
	// xrpc transport, a wrapper around http server
	client      *xrpc.Client
	handle      string
	apikey      string
	lastCreate  time.Time
	lastRefresh time.Time
	logger      *slog.Logger
}

// Creates new BlueSky Agent
func NewAgent(ctx context.Context, server string, handle string, apikey string) BskyAgent {

	if server == "" {
		server = defaultPDS
	}

	return BskyAgent{
		client: &xrpc.Client{
			Client: new(http.Client),
			Host:   server,
		},
		handle: handle,
		apikey: apikey,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

}

func (c *BskyAgent) WithLogger(l *slog.Logger) *BskyAgent {
	c.logger = l
	return c
}

func (c *BskyAgent) UploadImages(ctx context.Context, images ...Image) ([]lexutil.LexBlob, error) {
	c.Authenticate(ctx)

	for _, img := range images {
		getImage, err := getImageAsBuffer(img.Uri.String())
		if err != nil {
			log.Printf("Couldn't retrive the image: %v , %v", img, err)
		}

		resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(getImage))
		if err != nil {
			return nil, err
		}

		blob = append(blob, lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: resp.Blob.MimeType,
			Size:     resp.Blob.Size,
		})
	}
	return blob, nil
}

func (c *BskyAgent) UploadImage(ctx context.Context, image Image) (*lexutil.LexBlob, error) {
	c.Authenticate(ctx)

	getImage, err := getImageAsBuffer(image.Uri.String())
	if err != nil {
		log.Printf("Couldn't retrive the image: %v , %v", image, err)
	}

	resp, err := atproto.RepoUploadBlob(ctx, c.client, bytes.NewReader(getImage))
	if err != nil {
		return nil, err
	}

	blob := lexutil.LexBlob{
		Ref:      resp.Blob.Ref,
		MimeType: resp.Blob.MimeType,
		Size:     resp.Blob.Size,
	}

	return &blob, nil
}

// Post to social app
func (c *BskyAgent) PostToFeed(ctx context.Context, post appbsky.FeedPost) (string, string, error) {
	c.Authenticate(ctx)

	post_input := &atproto.RepoCreateRecord_Input{
		// collection: The NSID of the record collection.
		Collection: "app.bsky.feed.post",
		// repo: The handle or DID of the repo (aka, current account).
		Repo: c.client.Auth.Did,
		// record: The record itself. Must contain a $type field.
		Record: &lexutil.LexiconTypeDecoder{Val: &post},
	}

	response, err := atproto.RepoCreateRecord(ctx, c.client, post_input)
	if err != nil {
		return "", "", fmt.Errorf("unable to post, %v", err)
	}

	return response.Cid, response.Uri, nil
}

func getImageAsBuffer(imageURL string) ([]byte, error) {
	// Fetch image
	response, err := http.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Check response status
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", response.Status)
	}

	// Read response body
	imageData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}
