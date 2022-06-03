package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func main() {
	exitCode := run(os.Args, os.Stdin, os.Stdout, os.Stderr)
	os.Exit(exitCode)
}

func run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	program := args[0]

	flagSet := flag.NewFlagSet(program, flag.ContinueOnError)
	flagSet.SetOutput(stderr)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Usage of mddiffcheck:")
		fmt.Fprintln(flagSet.Output(), "  mddiffcheck -repo=<repo> <markdown-file> [markdown-file] ...")
		fmt.Fprintln(flagSet.Output(), "")
		fmt.Fprintln(flagSet.Output(), "Example:")
		fmt.Fprintln(flagSet.Output(), "  mddiffcheck -repo=https://github.com/user/repo doc1.md doc2.md")
		fmt.Fprintln(flagSet.Output(), "")
		fmt.Fprintln(flagSet.Output(), "Flags:")
		flagSet.PrintDefaults()
	}

	flagHelp := flagSet.Bool("help", false, "print this help")
	flagRepo := flagSet.String("repo", "", "repository to verify diffs against")
	err := flagSet.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(stderr, "%v\n", err)
		return 2
	}

	posArgs := flagSet.Args()
	if *flagHelp || len(posArgs) == 0 {
		flagSet.Usage()
		return 0
	}

	mdPaths := posArgs

	err = checkFiles(stderr, *flagRepo, mdPaths)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}

	return 0
}

func checkFiles(stderr io.Writer, repo string, mdPaths []string) error {
	workingDir, err := os.MkdirTemp("", "")
	if err != nil {
		return fmt.Errorf("making temp dir for repo clone: %w", err)
	}

	fmt.Fprintf(stderr, "repo: cloning %s into %s...\n", repo, workingDir)
	err = gitClone(workingDir, repo)
	if err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	fmt.Fprintf(stderr, "repo: ok\n")

	errored := false
	for _, mdPath := range mdPaths {
		mdFile, err := os.Open(mdPath)
		if err != nil {
			return err
		}
		defer mdFile.Close()

		err = checkFile(stderr, workingDir, mdPath, mdFile)
		if err != nil {
			errored = true
		}
	}

	if errored {
		return fmt.Errorf("one or more diffs failed to apply")
	}
	return nil
}

func checkFile(stderr io.Writer, workingDir, filename string, markdown io.Reader) error {
	checkDiff := func(lineNum int, params, diff string) error {
		fmt.Fprintf(stderr, "%s:%d: parsing diff\n", filename, lineNum)

		paramValues, err := url.ParseQuery(params)
		if err != nil {
			return fmt.Errorf("parsing params %q: %w", params, err)
		}

		ignore := paramValues.Get("mddiffcheck.ignore")
		if ignore == "true" {
			fmt.Fprintf(stderr, "%s:%d: skipping due to mddiffcheck.ignore=true\n", filename, lineNum)
			return nil
		}

		base := paramValues.Get("mddiffcheck.base")
		if base == "" {
			return fmt.Errorf("no base specified for diff")
		}
		fmt.Fprintf(stderr, "%s:%d: checking out base ref %s\n", filename, lineNum, base)
		err = gitFetch(workingDir, base)
		if err != nil {
			return fmt.Errorf("fetching %q: %w", base, err)
		}
		err = gitCheckout(workingDir, base)
		if err != nil {
			return fmt.Errorf("checkout out ref %q: %w", base, err)
		}

		diffFile, err := os.CreateTemp("", "")
		if err != nil {
			return fmt.Errorf("creating temp diff file: %w", err)
		}
		_, err = diffFile.WriteString(diff)
		if err != nil {
			return fmt.Errorf("writing temp diff file: %w", err)
		}
		err = diffFile.Close()
		if err != nil {
			return fmt.Errorf("closing temp diff file: %w", err)
		}
		fmt.Fprintf(stderr, "%s:%d: checking diff (%s)...\n", filename, lineNum, diffFile.Name())
		err = gitApply(workingDir, diffFile.Name())
		if err != nil {
			return fmt.Errorf("git apply: %w", err)
		}
		fmt.Fprintf(stderr, "%s:%d: ok\n", filename, lineNum)
		return nil
	}

	printErr := func(f findDiffsCallback) findDiffsCallback {
		return func(lineNum int, params, diff string) error {
			err := f(lineNum, params, diff)
			if err != nil {
				fmt.Fprintf(stderr, "%s:%d: error: %v\n", filename, lineNum, err)
			}
			return err
		}
	}

	err := findDiffs(markdown, printErr(checkDiff))
	if err != nil {
		return err
	}

	return nil
}

type findDiffsCallback func(lineNum int, commit, diff string) error

func findDiffs(md io.Reader, found findDiffsCallback) error {
	p := goldmark.DefaultParser()

	b, err := ioutil.ReadAll(md)
	if err != nil {
		return fmt.Errorf("reading markdown file: %w", err)
	}

	rootNode := p.Parse(text.NewReader(b))
	return ast.Walk(rootNode, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() != ast.KindFencedCodeBlock {
			return ast.WalkContinue, nil
		}
		fcb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Ignore empty code blocks.
		if fcb.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}

		if fcb.Info == nil {
			return ast.WalkContinue, nil
		}
		info := strings.SplitN(string(fcb.Info.Text(b)), " ", 2)

		// Ignore code blocks that aren't diffs
		if info[0] != "diff" {
			return ast.WalkContinue, nil
		}

		// Get the params that appear after the word "diff" in the info.
		params := ""
		if len(info) >= 2 {
			params = strings.ReplaceAll(info[1], " ", "&")
		}

		// Get the line number the diff starts on.
		firstLine := fcb.Lines().At(0)
		lineNum := bytes.Count(b[:firstLine.Start], []byte("\n"))

		// Get the diff.
		diff := strings.Builder{}
		for i := 0; i < fcb.Lines().Len(); i++ {
			line := fcb.Lines().At(i)
			lineText := line.Value(b)
			diff.Write(lineText)
		}

		err = found(lineNum, params, diff.String())
		return ast.WalkSkipChildren, err
	})
}

func gitClone(dir, repo string) error {
	cmd := exec.Command("git", "clone", repo, dir)
	out := strings.Builder{}
	cmd.Stderr = &out
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("cloning %s into %s: %v\n%s", repo, dir, err, strings.TrimSuffix(out.String(), "\n"))
	}
	return nil
}

func gitFetch(dir, ref string) error {
	cmd := exec.Command("git", "fetch", "--quiet", "origin", ref)
	out := strings.Builder{}
	cmd.Stderr = &out
	cmd.Stdout = &out
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("fetch out %s: %v\n%s", ref, err, strings.TrimSuffix(out.String(), "\n"))
	}
	return nil
}

func gitCheckout(dir, ref string) error {
	cmd := exec.Command("git", "checkout", "--quiet", ref)
	out := strings.Builder{}
	cmd.Stderr = &out
	cmd.Stdout = &out
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("checkout out %s into %s: %v\n%s", ref, dir, err, strings.TrimSuffix(out.String(), "\n"))
	}
	return nil
}

func gitApply(dir, file string) error {
	cmd := exec.Command("git", "apply", "--check", file)
	out := strings.Builder{}
	cmd.Stderr = &out
	cmd.Stdout = &out
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("applying %s into %s: %v\n%s", file, dir, err, strings.TrimSuffix(out.String(), "\n"))
	}
	return nil
}
