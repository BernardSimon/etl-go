package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSourceSQL_AllowsSelect(t *testing.T) {
	assert.NoError(t, ValidateSourceSQL("SELECT * FROM users"))
	assert.NoError(t, ValidateSourceSQL("select id from users where id = 1"))
	assert.NoError(t, ValidateSourceSQL("  SELECT count(*) FROM orders"))
	assert.NoError(t, ValidateSourceSQL("WITH cte AS (SELECT 1) SELECT * FROM cte"))
}

func TestValidateSourceSQL_RejectsNonSelect(t *testing.T) {
	assert.Error(t, ValidateSourceSQL("INSERT INTO users VALUES (1)"))
	assert.Error(t, ValidateSourceSQL("UPDATE users SET name = 'x'"))
	assert.Error(t, ValidateSourceSQL("DELETE FROM users"))
	assert.Error(t, ValidateSourceSQL("DROP TABLE users"))
	assert.Error(t, ValidateSourceSQL(""))
}

func TestValidateExecutorSQL_AllowsNormal(t *testing.T) {
	assert.NoError(t, ValidateExecutorSQL("INSERT INTO users VALUES (1)", false))
	assert.NoError(t, ValidateExecutorSQL("UPDATE users SET name = 'x'", false))
	assert.NoError(t, ValidateExecutorSQL("DELETE FROM users WHERE id = 1", false))
}

func TestValidateExecutorSQL_RejectsDangerous(t *testing.T) {
	assert.Error(t, ValidateExecutorSQL("DROP TABLE users", false))
	assert.Error(t, ValidateExecutorSQL("TRUNCATE TABLE users", false))
	assert.Error(t, ValidateExecutorSQL("ALTER TABLE users ADD col int", false))
}

func TestValidateExecutorSQL_AllowsDangerousWhenExplicit(t *testing.T) {
	assert.NoError(t, ValidateExecutorSQL("DROP TABLE users", true))
	assert.NoError(t, ValidateExecutorSQL("TRUNCATE TABLE users", true))
}

func TestValidateExecutorSQL_Empty(t *testing.T) {
	assert.Error(t, ValidateExecutorSQL("", false))
}
