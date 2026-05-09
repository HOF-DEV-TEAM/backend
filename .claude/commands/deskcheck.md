# /deskcheck — Desk Check / QA

## Instructions for Claude

When this command runs, you must:

1. **Identify what changed.** Run `git diff master...HEAD --name-only` and read every changed file.

2. **Trace the blast radius.** For each changed file, follow its usage through the stack:
   - Domain entity or repository changed → find every service method that calls it → find every handler that calls that service → those are the endpoints to test.
   - Application service changed → find every handler that calls it → test those endpoints.
   - Handler changed → test that endpoint directly.
   - Infrastructure (persistence, mailer, storage) changed → find the service(s) that depend on it → trace up to the handler.
   **Only test endpoints reachable from at least one changed file through this chain. Do not test unrelated endpoints.**

3. **Use curl exclusively** — no Go test files, no Postman, no browser.

4. **Use the fixed credentials and device constants defined below** — do not generate random emails, timestamps, or UUIDs.

5. **Test against the local server** (`http://localhost:8080`) — no Docker, no staging.

6. After every curl call — whether the expected result is a success or a deliberate failure — print:
   - **Request:** method, URL, headers (excluding the full token — show only the first 20 chars), and the full request body/params.
   - **Response:** the full raw response body.
   - **Verdict:** PASS or FAIL against the stated expectation, with a one-line reason if FAIL.

7. At the end, compile the full Score table (PASS / FAIL per endpoint) and post it as a GitHub PR comment using but ask first:
   ```bash
   gh pr comment <PR_NUMBER> --body "$(cat <<'BODY'
   ## Deskcheck Results

   | Endpoint | Scenario | Expected | Result |
   |----------|----------|----------|--------|
   | POST /session/sign_in | known device re-login | device count unchanged | PASS |
   | ... | ... | ... | ... |

   **Tested:** <date> · **Server:** localhost:8080 · **Branch:** <branch>
   BODY
   )"
   ```
   Derive the PR number from `gh pr view --json number -q .number`. Fill every row from the actual curl results — do not fabricate any entry.

---

## Constants — use these throughout every test

```bash
BASE=http://localhost:8080

# Fixed admin (created via: make seed-admin EMAIL=admin@hofng.org FIRST=Admin LAST=HOF PASS=AdminPass123!)
ADMIN_EMAIL="admin@hofng.org"
ADMIN_PASS="AdminPass123!"

# Fixed member user (created once in Setup below, reused every run)
USER_EMAIL="deskcheck@hofng.org"
USER_PASS="DeskCheck123!"

# Fixed devices — used in all device tests
DEVICE_A='{"who":"DeskCheck","identifier":"deskcheck-pixel-001","oas":"Android","brand":"Google Pixel","version":"14"}'
DEVICE_B='{"who":"DeskCheck","identifier":"deskcheck-iphone-002","os":"iOS","brand":"Apple iPhone","version":"17"}'
```

---

## Setup

```bash
# 1. Confirm server is running locally
curl -s $BASE/health
# Expected: {"status":"ok"}

# If not running: make build && ./bin/server &

# 2. Create the fixed member user (idempotent — 409 on second run is fine)
curl -s -X POST $BASE/session/sign_up \
  -H "Content-Type: application/json" \
  -d "{\"first_name\":\"Desk\",\"last_name\":\"Check\",\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"devices\":[$DEVICE_A]}"
# Expected: success:true (first run) or 409 (already exists — both OK)

# 3. Sign in as admin
ADMIN_SIGNIN=$(curl -s -X POST $BASE/session/sign_in/admin \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASS\"}")
ADMIN_TOKEN=$(echo "$ADMIN_SIGNIN" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "ADMIN_TOKEN=${ADMIN_TOKEN:0:40}..."
# Expected: token present

# 4. Sign in as member
USER_SIGNIN=$(curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"device\":$DEVICE_A}")
USER_TOKEN=$(echo "$USER_SIGNIN" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "USER_TOKEN=${USER_TOKEN:0:40}..."
# Expected: token present
```

---

## 1 — Device upsert on login

Verifies that every login upserts (not duplicates) the device record.

```bash
# Baseline: list devices — should have exactly 1 entry (DEVICE_A from signup)
curl -s $BASE/user/devices/all -H "Authorization: Bearer $USER_TOKEN" | grep -o '"identifier":"[^"]*"'
# Expected: one entry with identifier="deskcheck-pixel-001"

# Sign in again with the SAME device — should update, not duplicate
curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"device\":$DEVICE_A}" > /dev/null

USER_TOKEN=$(curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"device\":$DEVICE_A}" \
  | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)

curl -s $BASE/user/devices/all -H "Authorization: Bearer $USER_TOKEN" | grep -o '"identifier":"[^"]*"'
# Expected: STILL only one entry for "deskcheck-pixel-001" (no duplicate)

# Sign in with a NEW device — should append
NEW_SIGNIN=$(curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"device\":$DEVICE_B}")
USER_TOKEN=$(echo "$NEW_SIGNIN" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)

curl -s $BASE/user/devices/all -H "Authorization: Bearer $USER_TOKEN" | grep -o '"identifier":"[^"]*"'
# Expected: TWO entries — "deskcheck-pixel-001" and "deskcheck-iphone-002"

# Sign in with DEVICE_B again — no further duplicates
curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"device\":$DEVICE_B}" > /dev/null

curl -s $BASE/user/devices/all -H "Authorization: Bearer $USER_TOKEN" | grep -o '"identifier":"[^"]*"'
# Expected: STILL two entries (idempotent)
```

---

## 2 — Access level enforcement on audio messages

Verifies that members cannot access stewards/leaders-only content.

```bash
# Create a members-level message (default, visible to all)
MSG_MEMBERS=$(curl -s -X POST $BASE/admin/audio_message \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Members Sermon","author":"Pastor","description":"For all","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":true,"access":"members"}')
MSG_MEMBERS_ID=$(echo "$MSG_MEMBERS" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "MSG_MEMBERS_ID=$MSG_MEMBERS_ID"
# Expected: success:true

# Create a stewards-level message (restricted)
MSG_STEWARDS=$(curl -s -X POST $BASE/admin/audio_message \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Stewards Sermon","author":"Pastor","description":"Stewards only","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":false,"access":"stewards"}')
MSG_STEWARDS_ID=$(echo "$MSG_STEWARDS" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "MSG_STEWARDS_ID=$MSG_STEWARDS_ID"
# Expected: success:true

# Member gets members-level message → 200
curl -s $BASE/audio_message/id/message/$MSG_MEMBERS_ID \
  -H "Authorization: Bearer $USER_TOKEN"
# Expected: success:true, message returned

# Member gets stewards-level message → 403
curl -s $BASE/audio_message/id/message/$MSG_STEWARDS_ID \
  -H "Authorization: Bearer $USER_TOKEN"
# Expected: success:false, 403 Forbidden

# Admin gets stewards-level message → 200
curl -s $BASE/audio_message/id/message/$MSG_STEWARDS_ID \
  -H "Authorization: Bearer $ADMIN_TOKEN"
# Expected: success:true

# List as member with access=members filter — should NOT include stewards message
curl -s "$BASE/audio_message/?page=1&page_size=20&access=members" \
  -H "Authorization: Bearer $USER_TOKEN" | grep -o '"title":"[^"]*"'
# Expected: "Members Sermon" present; "Stewards Sermon" absent

# Cleanup
curl -s -X DELETE $BASE/admin/audio_message/delete/$MSG_MEMBERS_ID -H "Authorization: Bearer $ADMIN_TOKEN"
curl -s -X DELETE $BASE/admin/audio_message/delete/$MSG_STEWARDS_ID -H "Authorization: Bearer $ADMIN_TOKEN"
# Expected: success:true both
```

---

## 3 — ResetPassword OTP bypass prevention

Verifies that reset_password is blocked unless the OTP token has been validated first.

```bash
# Attempt reset WITHOUT going through forgot_password + verify_token → must fail
curl -s -X POST $BASE/user/reset_password \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"password_confirm\":\"$USER_PASS\"}"
# Expected: success:false, 403 (no validated OTP exists)

# Trigger forgot_password to generate an OTP (email queued — check DB below)
curl -s -X POST $BASE/session/forgot_password \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\"}"
# Expected: success:true

# Try reset again — OTP exists but NOT validated yet → still 403
curl -s -X POST $BASE/user/reset_password \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\",\"password\":\"$USER_PASS\",\"password_confirm\":\"$USER_PASS\"}"
# Expected: success:false, 403 (token.Validated is false)

# Verify OTP with wrong code → 400/403
curl -s -X PUT $BASE/session/verify_token \
  -H "Content-Type: application/json" \
  -d "{\"target\":\"$USER_EMAIL\",\"token\":\"000000\"}"
# Expected: success:false (wrong OTP)
```

---

## 4 — VerifyOTP response contains no password hash

Verifies the response is a DTO, not a raw domain User.

```bash
# Trigger a new OTP
curl -s -X POST $BASE/session/forgot_password \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\"}" > /dev/null

# Call verify_token with wrong OTP and inspect response fields
VERIFY_RESP=$(curl -s -X PUT $BASE/session/verify_token \
  -H "Content-Type: application/json" \
  -d "{\"target\":\"$USER_EMAIL\",\"token\":\"000000\"}")
echo "$VERIFY_RESP"

# Check: "password" must NOT appear anywhere in the response body
echo "$VERIFY_RESP" | grep -i '"password"'
# Expected: NO output (empty) — password field absent from response
```

---

## 5 — Async email queue (email_jobs table)

Verifies that emails are persisted to the DB rather than sent synchronously.

```bash
# Trigger an email — forgot_password queues a password-reset email
curl -s -X POST $BASE/session/forgot_password \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USER_EMAIL\"}"
# Expected: success:true (returns immediately — send is async)

# Query the email_jobs table to confirm a row was created
# Run this in psql or your DB shell:
#   SELECT id, "to", subject, status, attempts, scheduled_at
#   FROM email_jobs
#   WHERE "to" = 'deskcheck@hofng.org'
#   ORDER BY created_at DESC
#   LIMIT 5;
# Expected: at least one row with status='pending' or status='sent'
```

---

## 6 — Admin bootstrap: public sign_up/admin removed

Verifies that the old public admin signup endpoint no longer exists.

```bash
curl -s -X POST $BASE/session/sign_up/admin \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Hacker","last_name":"Test","email":"hack@example.com","password":"password123"}'
# Expected: 404 or 405 — route does not exist
```

---

## Score

| #  | Check                                        | Expected result                     |
|----|----------------------------------------------|-------------------------------------|
| 1a | Re-login with same device                    | Device count stays at 1, no dup     |
| 1b | Login with new device                        | Device count increases by 1         |
| 1c | Login with known device again                | Count stays the same (idempotent)   |
| 2a | Member reads members-level message           | 200 success                         |
| 2b | Member reads stewards-level message          | 403 Forbidden                       |
| 2c | Admin reads stewards-level message           | 200 success                         |
| 2d | List with access=members filter              | Stewards message absent             |
| 3a | Reset without any OTP                        | 403                                 |
| 3b | Reset with unvalidated OTP                   | 403                                 |
| 3c | Verify with wrong OTP                        | 400 or 403                          |
| 4  | VerifyOTP response has no `password` field   | Empty grep output                   |
| 5  | email_jobs row created after forgot_password | Row present in DB                   |
| 6  | POST /session/sign_up/admin                  | 404 or 405                          |

Any unexpected 500 is a bug — check server logs immediately.
