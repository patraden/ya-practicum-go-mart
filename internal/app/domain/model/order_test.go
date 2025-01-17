package model_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/patraden/ya-practicum-go-mart/internal/app/domain/model"
)

func TestOrderCheckLuhn(t *testing.T) {
	t.Parallel()

	order := model.NewOrder(567890123456, uuid.Nil)
	assert.True(t, order.CheckLuhn())

	order = model.NewOrder(456789123456, uuid.Nil)
	assert.True(t, order.CheckLuhn())

	order = model.NewOrder(224865435574328, uuid.Nil)
	assert.True(t, order.CheckLuhn())

	order = model.NewOrder(605650248566, uuid.Nil)
	assert.True(t, order.CheckLuhn())
}
