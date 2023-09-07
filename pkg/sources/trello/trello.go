package trello

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/source_metadatapb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/pb/sourcespb"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
)

const baseURL = "https://api.trello.com/1/"

type Source struct {
	name     string
	apiKey   string
	token    string
	boardIDs []string
	sourceId int64
	jobId    int64
	verify   bool
	jobPool  *errgroup.Group
	sources.Progress
	client *http.Client
	sources.CommonSourceUnitUnmarshaller
}

// Type returns the type of the source.
func (s *Source) Type() sourcespb.SourceType {
	return sourcespb.SourceType_SOURCE_TYPE_TRELLO
}

// Init returns an initialized Source.
func (s *Source) Init(_ context.Context, name string, jobId, sourceId int64, verify bool, connection *anypb.Any, concurrency int) error {
	s.name = name
	s.sourceId = sourceId
	s.jobId = jobId
	s.verify = verify
	s.jobPool = &errgroup.Group{}
	s.jobPool.SetLimit(concurrency)
	s.client = common.RetryableHttpClientTimeout(3)

	var conn sourcespb.Trello
	if err := anypb.UnmarshalTo(connection, &conn, proto.UnmarshalOptions{}); err != nil {
		return errors.WrapPrefix(err, "error unmarshalling connection", 0)
	}

	s.apiKey = conn.GetAuth().GetApiKey()
	s.token = conn.GetAuth().GetToken()
	s.boardIDs = conn.GetBoards()

	return nil
}

// Chunks emits chunks of bytes over a channel.
func (s *Source) Chunks(ctx context.Context, chunksChan chan *sources.Chunk) error {
	scanErrs := sources.NewScanErrors()

	for _, boardID := range s.boardIDs {
		board, err := s.getBoard(ctx, boardID)
		if err != nil {
			scanErrs.Add(err)
			return nil
		}

		if err = s.chunkBoard(ctx, board, chunksChan); err != nil {
			scanErrs.Add(err)
			return nil
		}
	}

	_ = s.jobPool.Wait()
	if scanErrs.Count() > 0 {
		ctx.Logger().V(2).Info("encountered errors while scanning", "count", scanErrs.Count(), "errors", scanErrs)
	}

	return nil
}

type board struct {
    // Your board's struct definition
}

func (s *Source) getBoard(_ context.Context, boardID string) (*board, error) {
	// Your APi Call Logic for fetching the board by the given ID
}

func (s *Source) chunkBoard(ctx context.Context, board *board, chunksChan chan *sources.Chunk) error {
	// Your logic for chunking a board
}
