package main

import (
	"path/filepath"
	"regexp"
	"testing"

	"4d63.com/testcli"
	"github.com/stretchr/testify/assert"
)

func TestCheckSuccess(t *testing.T) {
	h := testcli.Helper{TB: t}

	markdownPath := filepath.Join("testdata", "cap-0039.md")

	args := []string{"mddiffcheck", "-repo", "https://github.com/stellar/stellar-core", markdownPath}

	exitCode, stdout, stderr := h.Main(args, nil, run)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "", stdout)
	stderrSimplified := regexp.MustCompile(`/(tmp|var)([/_a-zA-Z0-9]*)/\d+`).ReplaceAllLiteralString(stderr, "/tmp/out")
	assert.Equal(
		t,
		`repo: cloning https://github.com/stellar/stellar-core into /tmp/out...
repo: ok
testdata/cap-0039.md:64: parsing diff
testdata/cap-0039.md:64: checking out base ref b9e10051eafa1125e8d238a47e5915dad30c2640
testdata/cap-0039.md:64: checking diff (/tmp/out)...
testdata/cap-0039.md:64: ok
`,
		stderrSimplified,
	)
}

func TestCheckFail(t *testing.T) {
	h := testcli.Helper{TB: t}

	markdownPath := filepath.Join("testdata", "cap-0039-fail.md")

	args := []string{"mddiffcheck", "-repo", "https://github.com/stellar/stellar-core", markdownPath}

	exitCode, stdout, stderr := h.Main(args, nil, run)
	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "", stdout)
	stderrSimplified := regexp.MustCompile(`/(tmp|var)([/_a-zA-Z0-9]*)/\d+`).ReplaceAllLiteralString(stderr, "/tmp/out")
	assert.Equal(
		t,
		`repo: cloning https://github.com/stellar/stellar-core into /tmp/out...
repo: ok
testdata/cap-0039-fail.md:64: parsing diff
testdata/cap-0039-fail.md:64: checking out base ref b9e10051eafa1125e8d238a47e5915dad30c2640
testdata/cap-0039-fail.md:64: checking diff (/tmp/out)...
testdata/cap-0039-fail.md:64: error: git apply: applying /tmp/out into /tmp/out: exit status 1
error: patch failed: src/xdr/Stellar-ledger-entries.x:114
error: src/xdr/Stellar-ledger-entries.x: patch does not apply
error: one or more diffs failed to apply
`,
		stderrSimplified,
	)
}

