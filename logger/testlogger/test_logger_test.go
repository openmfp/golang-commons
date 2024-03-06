package testlogger

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/openmfp/golang-commons/logger"
)

func TestTestLogger(t *testing.T) {
	t.Run("Messages", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			messages, err := testLogger.GetLogMessages()

			assert.NoError(t, err)
			assert.Equal(t, len(messages), 0)
		})

		t.Run("two entries", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			testLogger.Logger.Info().Msg("foo")
			testLogger.Logger.Debug().Msg("bar")

			messages, err := testLogger.GetLogMessages()

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, 2, len(messages))
			assert.Equal(t, zerolog.InfoLevel, messages[0].Level)
			assert.Equal(t, zerolog.DebugLevel, messages[1].Level)
		})

		// testcase for custom attributes in a log message
		t.Run("custom attributes", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			testLogger.Logger.Info().
				Str("customStr", "attribute").
				Int("customInt", 1).
				Interface("customInterface", 3).
				Msg("foo")

			messages, err := testLogger.GetLogMessages()

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, len(messages), 1)
			assert.Equal(t, zerolog.InfoLevel, messages[0].Level)
			assert.Equal(t, "foo", messages[0].Message)
			assert.Equal(t, (*string)(nil), messages[0].Error)
			assert.Equal(t, "attribute", messages[0].Attributes["customStr"])
			assert.Equal(t, float64(1), messages[0].Attributes["customInt"])
			assert.Equal(t, float64(3), messages[0].Attributes["customInterface"])
		})
	})

	t.Run("Messages for level", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			messages, err := testLogger.GetMessagesForLevel(logger.Level(zerolog.WarnLevel))
			testLogger.Logger.Info().Msg("foo")

			assert.NoError(t, err)
			assert.Equal(t, len(messages), 0)
		})

		t.Run("two warning", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			testLogger.Logger.Info().Msg("foo")
			testLogger.Logger.Debug().Msg("bar")
			testLogger.Logger.Warn().Msg("oh no")
			testLogger.Logger.Warn().Msg("two warnings")

			messages, err := testLogger.GetMessagesForLevel(logger.Level(zerolog.WarnLevel))

			assert.NoError(t, err)
			assert.Equal(t, len(messages), 2)
			assert.Equal(t, messages[0].Message, "oh no")
			assert.Equal(t, messages[1].Message, "two warnings")
		})
	})

	t.Run("Messages error messages", func(t *testing.T) {
		t.Run("empty", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			messages, err := testLogger.GetErrorMessages()
			testLogger.Logger.Info().Msg("foo")

			assert.NoError(t, err)
			assert.Equal(t, len(messages), 0)
		})

		t.Run("two errors", func(t *testing.T) {
			// Arrange
			testLogger := New()

			// Act
			testLogger.Logger.Info().Msg("foo")
			testLogger.Logger.Debug().Msg("bar")
			testLogger.Logger.Error().Msg("oh no")
			testLogger.Logger.Error().Msg("two errors")

			messages, err := testLogger.GetErrorMessages()

			assert.NoError(t, err)
			assert.Equal(t, len(messages), 2)
			assert.Equal(t, messages[0].Message, "oh no")
			assert.Equal(t, messages[1].Message, "two errors")
		})
	})
}
