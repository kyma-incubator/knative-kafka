package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test The GenerateValidDnsName() Functionality
func TestGenerateValidDnsName(t *testing.T) {

	// Define The TestCase Struct
	type TestCase struct {
		Name   string
		Length int
		Result string
	}

	// Create The TestCases
	testCases := []TestCase{
		{Name: "testname", Length: 10, Result: "testname"},
		{Name: "TeStNaMe", Length: 10, Result: "testname"},
		{Name: "testnamelooooong", Length: 10, Result: "testnamelo"},
		{Name: "testname1234567890", Length: 50, Result: "testname"},
		{Name: "abcdefghijk1234567890lmnopqrstuvwxyz", Length: 50, Result: "abcdefghijk1234567890lmnopqrstuvwxyz"},
		{Name: "a~!@#$%^&*()_+=`<>?/\\z", Length: 50, Result: "az"},
		{Name: "123testname", Length: 50, Result: "kk-123testname"},
		{Name: "123testNAME", Length: 10, Result: "kk-123test"},
		{Name: "a123456789012345678901234567890123456789012345678901234567890z", Length: -1, Result: "a123456789012345678901234567890123456789012345678901234567890z"},
		{Name: "a123456789012345678901234567890123456789012345678901234567890z", Length: 0, Result: "a123456789012345678901234567890123456789012345678901234567890z"},
		{Name: "abc123456789012345678901234567890123456789012345678901234567890z", Length: 3, Result: "abc"},
		{Name: "a123456789012345678901234567890123456789012345678901234567890z", Length: 100, Result: "a123456789012345678901234567890123456789012345678901234567890z"},
		{Name: "a123456789012345678901234567890123456789012345678901234567890zxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", Length: 100, Result: "a123456789012345678901234567890123456789012345678901234567890zx"},
	}

	// Run The TestCases
	for _, testCase := range testCases {
		actualDnsName := GenerateValidDnsName(testCase.Name, testCase.Length)
		assert.Equal(t, testCase.Result, actualDnsName)
	}
}
