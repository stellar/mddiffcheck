# mddiffcheck
A tool for checking that git diffs that have been in markdown files apply successfully to the repo they're intended.

This tool was created to verify that git diffs included in [stellar-protocol] Core Advancement Protocols (CAPs) are valid.

The tool runs `git` as a subprocess and requires it to be installed.

[stellar-protocol]: https://github.com/stellar/stellar-protocol

## Usage

```
$ go get github.com/stellar/mddiff@latest
```

```
$ mddiff -help
Usage of mddiffcheck:
  mddiffcheck -repo=<repo> <markdown-file> [markdown-file] ...

Example:
  mddiffcheck -repo=https://github.com/user/repo doc1.md doc2.md

Flags:
  -help
        print this help
  -repo string
        repository to verify diffs against
```

When adding diffs into markdown, use the following format:

- Specify the code block is a diff with `diff` on the same line as the backticks.
- Specify the base git reference the diff applies to, either a tag or a commit sha, using the `mddiffcheck.base=` parameter.
- If you want to ignore a specific diff code block you can use the `mddiffcheck.ignore=true` parameter.

````
# Heading

## Subheading

```diff mddiffcheck.base=v16.0.0
diff --git a/src/xdr/Stellar-ledger-entries.x b/src/xdr/Stellar-ledger-entries.x
index 0e7bc842..68c52758 100644
--- a/src/xdr/Stellar-ledger-entries.x
+++ b/src/xdr/Stellar-ledger-entries.x
@@ -114,12 +114,15 @@ enum AccountFlags
     // Trustlines are created with clawback enabled set to "true",
     // and claimable balances created from those trustlines are created
     // with clawback enabled set to "true"
-    AUTH_CLAWBACK_ENABLED_FLAG = 0x8
+    AUTH_CLAWBACK_ENABLED_FLAG = 0x8,
+    // Trustlines are created with revocation disabled set to "true"
+    AUTH_NOT_REVOCABLE_FLAG = 0x10
 };
 
 // mask for all valid flags

```

````
