package objectstore

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

/**
 * For saving static copy of original LPA to S3
 */
type s3Client interface {
	Put(objectKey string, obj any) (*s3.PutObjectOutput, error)
	Get(objectKey string) (*s3.GetObjectOutput, error)
}

type StaticLpaStorage struct {
	client s3Client
}

func (sls *StaticLpaStorage) Save(lpa *shared.Lpa) error {
	// save a copy of the original to permanent storage,
	// but only if the key doesn't already exist
	objectKey := fmt.Sprintf("%s/donor-executed-lpa.json", lpa.Uid)
	_, err := sls.client.Get(objectKey)

	if err == nil {
		// no error and 200 => bad (object already exists)
		return fmt.Errorf("Could not save donor executed LPA as key %s already exists", objectKey)
	}

	// error which is a 404 => good (object should not already exist)
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		_, err = sls.client.Put(objectKey, lpa)
    }

    return err
}

func NewStaticLpaStorage(client s3Client) *StaticLpaStorage {
	return &StaticLpaStorage{
		client: client,
	}
}
