## Preamble

```
CAP: 0039
Title: Not Auth Revocable Trustlines
Working Group:
    Owner: Leigh McCulloch <@leighmcculloch>
    Authors: Leigh McCulloch <@leighmcculloch>
    Consulted: Tomer Weller <@tomerweller>, Siddharth Suresh <@sisuresh>
Status: Draft
Created: 2021-05-17
Discussion: TBD
Protocol version: TBD
```

## Simple Summary

This CAP addresses the following authorization semantics requirements:
- It should be clear and predictable to an asset holder if their assets are
revocable.
- It should be possible for issuer accounts to communicate their intent to
revoke without giving up the mutability of their asset.
- It should be possible for issuer accounts to make their assets usable with
contracts, such as payment channels, without making their accounts immutable.

## Working Group

This protocol change was authored by Leigh McCulloch, with input from the
consulted individuals mentioned at the top of this document.

## Motivation

Trustline authorization is an important feature of the Stellar protocol. It
allows issuers to handle various regulatory requirements. However, its current
behavior forces asset issuers to make a choice between immutable and predictable
for asset holders, or mutable and unpredictable for asset holders.

The current behavior makes all non-immutable assets revocable even if the asset
issuer does not have auth revocable flag cleared, and even if issuer has no
intention to ever revoke existing trustlines, and has not set immutable for
other reasons.

Most assets on the Stellar network are not immutable.

This prevents most assets from being used in contracts since the revocation of a
trustline can break a contract. This may prevent most assets from being used in payment channels, such as those described in [CAP-21].

### Goals Alignment

This CAP is aligned with the following Stellar Network Goals:

- The Stellar Network should make it clear for asset holders to understand the
trust relationship they have with asset issuers.
- The Stellar Network should enable predictable lock-up of funds in escrow
accounts for contracts, such as payment channels.

## Specification

### XDR Changes

This patch of XDR changes is based on the XDR files in commit
`b9e10051eafa1125e8d238a47e5915dad30c2640` of stellar-core.

```diff mddiffcheck.base=b9e10051eafa1125e8d238a47e5915dad30c2640
diff --git a/src/xdr/Stellar-ledger-entries.x b/src/xdr/Stellar-ledger-entries.x
index 0e7bc842..68c52758 100644
--- a/src/xdr/Stellar-ledger-entries.x
+++ b/src/xdr/Stellar-ledger-entries.x
@@ -114,12 +114,15 @@ enum AccountFlags
     // Trustlines are created with clawback enabled set to "true",
     // and claimable balances created from those trustlines are created
     // with clawback enabled set to "true"
-    AUTH_CLAWBACK_ENABLED_FLAG = 0x6
+    AUTH_CLAWBACK_ENABLED_FLAG = 0x8,
+    // Trustlines are created with revocation disabled set to "true"
+    AUTH_NOT_REVOCABLE_FLAG = 0x10
 };
 
 // mask for all valid flags
 const MASK_ACCOUNT_FLAGS = 0x7;
 const MASK_ACCOUNT_FLAGS_V17 = 0xF;
+const MASK_ACCOUNT_FLAGS_V18 = 0x1F;
 
 // maximum number of signers
 const MAX_SIGNERS = 20;
@@ -206,13 +209,16 @@ enum TrustLineFlags
     AUTHORIZED_TO_MAINTAIN_LIABILITIES_FLAG = 2,
     // issuer has specified that it may clawback its credit, and that claimable
     // balances created with its credit may also be clawed back
-    TRUSTLINE_CLAWBACK_ENABLED_FLAG = 4
+    TRUSTLINE_CLAWBACK_ENABLED_FLAG = 4,
+    // issuer has specified that it may not revoke authorization.
+    TRUSTLINE_NOT_REVOCABLE_FLAG = 8,
 };
 
 // mask for all trustline flags
 const MASK_TRUSTLINE_FLAGS = 1;
 const MASK_TRUSTLINE_FLAGS_V13 = 3;
 const MASK_TRUSTLINE_FLAGS_V17 = 7;
+const MASK_TRUSTLINE_FLAGS_V18 = 15;
 
 struct TrustLineEntry
 {

```

### Semantics

This proposal introduces one new account flag that controls whether new trustlines are created with revocation disabled or not.

This proposal introduces one new trustline flag that captures onto the trustline
at the moment it is created whether it will be revocable by the issuer account.

This proposal changes the `AllowTrustOp` to disallow its use to revoke or reduce
the limit of a trustline that has the new trustline flag set. This prevents
authorization revocation on trustlines when the issuer has indicated it does not
intend to revoke trustlines while allowing the issuer account to remain mutable.

Existing and new trustlines for issuers that do not use the new account flag are
unaffected and will use the existing behavior they have today. An issuer may
revoke existing trustline at anytime by enabling the `AUTH_REVOCABLE_FLAG`
account flag on the issuer account, and using the `AllowTrustOp` operation to
revoke authorization of the trustor.

New trustlines created when the issuer account has its `AUTH_NOT_REVOCABLE_FLAG`
will not be revocable, even if the issuer account sets the `AUTH_REVOCABLE_FLAG`
account flag at a later time.

#### Account Flags

This proposal introduces a new account flag:
- `AUTH_NOT_REVOCABLE_FLAG` that indicates if trustlines should be created
not revocable.

#### TrustLine Flags

This proposal introduces a new trustline flag that is set on the trustline
when it is created:
- `TRUSTLINE_NOT_REVOCABLE_FLAG` that is set if the issuer account has its
`AUTH_NOT_REVOCABLE_FLAG` flag set.

#### Change Trust Operation

This proposal introduces changes to the semantics of the `ChangeTrustOp`
operation.

When `ChangeTrustOp` creates a new trustline it sets 
`TRUSTLINE_NOT_REVOCABLE_FLAG` if the `AUTH_NOT_REVOCABLE_FLAG` account flag is
set on the issuer account.

When `ChangeTrustOp` modifies an existing trustline the new flag is not changed,
regardless of the state of the `AUTH_NOT_REVOCABLE_FLAG` or
`AUTH_REVOCABLE_FLAG` account flags on the issuer account.

#### Allow Trust Operation

This proposal introduces changes to the semantics of the `ALLOW_TRUST` operation.

- Disallow `ALLOW_TRUST` operations that downgrade authorization when the
trustline is authorized and the `TRUSTLINE_NOT_REVOCABLE_FLAG` trustline flag is
set.

#### Set TrustLine Flags Operation

This proposal extends the cases that the `SetTrustLineFlagsOp` operation will return `SET_TRUST_LINE_FLAGS_MALFORMED` during validation to include the following conditions:
- `TRUSTLINE_NOT_REVOCABLE_FLAG` is set on `clearFlags`.

Issuer accounts could use `SetTrustLineFlagsOp` to set the
`TRUSTLINE_NOT_REVOCABLE_FLAG` on accounts, making it such
that existing trustlines cannot be revoked.

## Design Rationale

The `ChangeTrustOp` semantics introduced are consistent with the semantics of the `TRUSTLINE_CLAWBACK_ENABLED_FLAG` flag that was introduced in [CAP-35].

A disabled flag is introduced because disabling revocation is the new behavior.
Existing trustlines of non-immutable issuers are already revocable enabled even
though no flag on the trustline indicates that.

A new not revocable flag is introduced onto accounts so that existing issuer accounts see no change in behavior.

## Protocol Upgrade Transition

### Backwards Incompatibilities

This proposal is backwards compatible with existing and new trustlines.

This proposal is backwards compatible for existing and new issuers who expect to
enable auth revocation in the future but do not wish to enable it today.

### Resource Utilization

No substantial changes to resource utilization.

## Test Cases

None yet.

## Security Concerns

None known.

## Implementation

None yet.

[CAP-21]: ./cap-0021.md
[CAP-35]: ./cap-0035.md
