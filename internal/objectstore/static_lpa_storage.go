package objectstore

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-data-lpa-store/internal/shared"
)

/**
 * For saving static copy of original LPA to S3
 */
type StaticLpaStorage struct {
	client *S3Client
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

func NewStaticLpaStorage(client *S3Client) *StaticLpaStorage {
	return &StaticLpaStorage{
		client: client,
	}
}
