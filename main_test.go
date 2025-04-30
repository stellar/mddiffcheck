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

	markdownPath1 := filepath.Join("testdata", "cap-0039.md")
	markdownPath2 := filepath.Join("testdata", "cap-0070.md")

	args := []string{
		"mddiffcheck",
		"-repo",
		"https://github.com/stellar/stellar-core https://github.com/stellar/stellar-xdr",
		markdownPath1,
		markdownPath2,
	}

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
testdata/cap-0039.md:64: fetched from https://github.com/stellar/stellar-core
testdata/cap-0039.md:64: checking diff (/tmp/out)...
testdata/cap-0039.md:64: ok
testdata/cap-0070.md:49: parsing diff
testdata/cap-0070.md:49: checking out base ref 8903b65de5cb56e361800e93aa339ab1a5c1a2e7
testdata/cap-0070.md:49: fetched from https://github.com/stellar/stellar-xdr
testdata/cap-0070.md:49: checking diff (/tmp/out)...
testdata/cap-0070.md:49: ok
`,
		stderrSimplified,
	)
}

func TestFetchFail(t *testing.T) {
	h := testcli.Helper{TB: t}

	markdownPath1 := filepath.Join("testdata", "cap-0039.md")
	markdownPath2 := filepath.Join("testdata", "cap-0070.md")

	args := []string{
		"mddiffcheck",
		"-repo",
		"https://github.com/stellar/stellar-core https://github.com/stellar/rs-stellar-xdr",
		markdownPath1,
		markdownPath2,
	}

	exitCode, stdout, stderr := h.Main(args, nil, run)
	assert.Equal(t, 1, exitCode)
	assert.Equal(t, "", stdout)
	stderrSimplified := regexp.MustCompile(`/(tmp|var)([/_a-zA-Z0-9]*)/\d+`).ReplaceAllLiteralString(stderr, "/tmp/out")
	assert.Equal(
		t,
		`repo: cloning https://github.com/stellar/stellar-core into /tmp/out...
repo: ok
testdata/cap-0039.md:64: parsing diff
testdata/cap-0039.md:64: checking out base ref b9e10051eafa1125e8d238a47e5915dad30c2640
testdata/cap-0039.md:64: fetched from https://github.com/stellar/stellar-core
testdata/cap-0039.md:64: checking diff (/tmp/out)...
testdata/cap-0039.md:64: ok
testdata/cap-0070.md:49: parsing diff
testdata/cap-0070.md:49: checking out base ref 8903b65de5cb56e361800e93aa339ab1a5c1a2e7
testdata/cap-0070.md:49: error: fetching "8903b65de5cb56e361800e93aa339ab1a5c1a2e7":
fetch 8903b65de5cb56e361800e93aa339ab1a5c1a2e7 from https://github.com/stellar/stellar-core: exit status 128
fatal: remote error: upload-pack: not our ref 8903b65de5cb56e361800e93aa339ab1a5c1a2e7
fetch 8903b65de5cb56e361800e93aa339ab1a5c1a2e7 from https://github.com/stellar/rs-stellar-xdr: exit status 128
fatal: remote error: upload-pack: not our ref 8903b65de5cb56e361800e93aa339ab1a5c1a2e7
error: one or more diffs failed to apply
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
testdata/cap-0039-fail.md:64: fetched from https://github.com/stellar/stellar-core
testdata/cap-0039-fail.md:64: checking diff (/tmp/out)...
testdata/cap-0039-fail.md:64: error: git apply: applying /tmp/out into /tmp/out: exit status 1
error: patch failed: src/xdr/Stellar-ledger-entries.x:114
error: src/xdr/Stellar-ledger-entries.x: patch does not apply
error: one or more diffs failed to apply
`,
		stderrSimplified,
	)
}

func TestCheckSuccessWithOrphanCommit(t *testing.T) {
	h := testcli.Helper{TB: t}

	markdownPath := filepath.Join("testdata", "cap-0048.md")

	args := []string{"mddiffcheck", "-repo", "https://github.com/stellar/stellar-core", markdownPath}

	exitCode, stdout, stderr := h.Main(args, nil, run)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "", stdout)
	stderrSimplified := regexp.MustCompile(`/(tmp|var)([/_a-zA-Z0-9]*)/\d+`).ReplaceAllLiteralString(stderr, "/tmp/out")
	assert.Equal(
		t,
		`repo: cloning https://github.com/stellar/stellar-core into /tmp/out...
repo: ok
testdata/cap-0048.md:79: parsing diff
testdata/cap-0048.md:79: checking out base ref 7fcc8002a595e59fad2c9bedbcf019865fb6b373
testdata/cap-0048.md:79: fetched from https://github.com/stellar/stellar-core
testdata/cap-0048.md:79: checking diff (/tmp/out)...
testdata/cap-0048.md:79: ok
`,
		stderrSimplified,
	)
}
