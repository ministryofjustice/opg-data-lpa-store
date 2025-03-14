package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountAttorneys(t *testing.T) {
	actives, replacements := CountAttorneys([]Attorney{}, []TrustCorporation{})
	assert.Equal(t, 0, actives)
	assert.Equal(t, 0, replacements)

	actives, replacements = CountAttorneys([]Attorney{
		{
			Status:          AttorneyStatusInactive,
			AppointmentType: AppointmentTypeReplacement,
		},
		{
			Status:          AttorneyStatusActive,
			AppointmentType: AppointmentTypeOriginal,
		},
		{
			Status:          AttorneyStatusInactive,
			AppointmentType: AppointmentTypeReplacement,
		},
	}, []TrustCorporation{
		{
			Status:          AttorneyStatusInactive,
			AppointmentType: AppointmentTypeReplacement,
		},
		{
			Status:          AttorneyStatusActive,
			AppointmentType: AppointmentTypeOriginal,
		},
	})
	assert.Equal(t, 2, actives)
	assert.Equal(t, 3, replacements)
}
