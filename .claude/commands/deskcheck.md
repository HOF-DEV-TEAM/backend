# /deskcheck — Desk Check / QA

Manual curl-based verification of one or all API endpoints.
Arguments: $ARGUMENTS — optional scope (e.g. "subscription", "user", "auth", or blank for all).

---

## Setup

```bash
BASE=http://localhost:8080

# Confirm server is running
curl -s $BASE/health
# Expected: {"status":"ok"}

# If server is not running:
make build && ./bin/server &
sleep 3
```

---

## Get a token

```bash
# Sign in to get TOKEN
SIGNIN=$(curl -s -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}')

TOKEN=$(echo "$SIGNIN" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4)
REFRESH=$(echo "$SIGNIN" | grep -o '"refresh_token":"[^"]*"' | head -1 | cut -d'"' -f4)

echo "TOKEN=${TOKEN:0:40}..."

# Verify global_parameters in session response
echo "$SIGNIN" | grep -o '"global_parameters":{[^}]*}'
```

---

## Auth / Session checks

```bash
# Sign up (no devices)
curl -s -X POST $BASE/session/sign_up \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"User","email":"test_$(date +%s)@example.com","password":"pass123456"}'

# Sign up with devices (array)
curl -s -X POST $BASE/session/sign_up \
  -H "Content-Type: application/json" \
  -d '{
    "first_name":"Multi","last_name":"Device","email":"multi_$(date +%s)@example.com","password":"pass123456",
    "devices":[
      {"who":"me","identifier":"dev-001","os":"Android","brand":"Samsung","version":"13"},
      {"who":"me","identifier":"dev-002","os":"iOS","brand":"Apple","version":"17"}
    ]
  }'
# Expected: success:true, user object with roles:["member"]

# Token refresh
curl -s -X POST $BASE/session/authenticate \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\",\"refresh_token\":\"$REFRESH\"}"
# Expected: new token + refresh_token

# Forgot password
curl -s -X POST $BASE/session/forgot_password \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com"}'
# Expected: success:true (email sent, or graceful failure if mailer unconfigured)

# Verify OTP
curl -s -X PUT $BASE/session/verify_token \
  -H "Content-Type: application/json" \
  -d '{"target":"admin@example.com","token":"123456"}'
# Expected: success:false (invalid OTP) — correct
```

---

## User checks

```bash
# Update profile
curl -s -X POST $BASE/user/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Updated","last_name":"Name"}'
# Expected: success:true

# Change password (correct old password)
curl -s -X POST $BASE/user/change_password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","old_password":"admin123","new_password":"admin123","confirm_new_password":"admin123"}'
# Expected: success:true

# Reset password
curl -s -X POST $BASE/user/reset_password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123","password_confirm":"admin123"}'
# Expected: success:true

# Send email verification
curl -s -X POST $BASE/user/verify_email \
  -H "Authorization: Bearer $TOKEN"
# Expected: success:true (or mailer error — non-blocking)

# Get roles
curl -s $BASE/user/roles -H "Authorization: Bearer $TOKEN"
# Expected: success:true, array of roles

# Devices — list
curl -s $BASE/user/devices/all -H "Authorization: Bearer $TOKEN"

# Devices — add one
curl -s -X POST $BASE/user/devices/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"who":"me","identifier":"dev-check-001","os":"Android","brand":"Pixel","version":"14"}'
# Expected: success:true

# Devices — delete
curl -s -X DELETE $BASE/user/devices/delete/dev-check-001 \
  -H "Authorization: Bearer $TOKEN"
# Expected: success:true

# Favourites — add
curl -s -X POST $BASE/user/favourite \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message_id":"00000000-0000-0000-0000-000000000000","fav":true}'
# Expected: success:true OR 404 (message not found — both acceptable)

# Favourites — list
curl -s $BASE/user/favourite/favs -H "Authorization: Bearer $TOKEN"
# Expected: success:true, array
```

---

## Content checks

```bash
# Create audio message
MSG=$(curl -s -X POST $BASE/audio_message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Desk Check Sermon","author":"Pastor","description":"Test","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":true}')
MSG_ID=$(echo "$MSG" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "MSG_ID=$MSG_ID"
# Expected: success:true, created message

# List messages
curl -s "$BASE/audio_message/?page=1&page_size=5" -H "Authorization: Bearer $TOKEN"
# Expected: success:true, array with total

# Get message
curl -s $BASE/audio_message/id/message/$MSG_ID -H "Authorization: Bearer $TOKEN"
# Expected: success:true, message object

# Update message
curl -s -X PUT $BASE/audio_message/update/$MSG_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Sermon","author":"Pastor","description":"Updated","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":false}'
# Expected: success:true

# Create meditation
MED=$(curl -s -X POST $BASE/audio_message/meditation \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Desk Check Meditation","image":"https://example.com/med.jpg","status":"active"}')
MED_ID=$(echo "$MED" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "MED_ID=$MED_ID"
# Expected: success:true

# List meditations
curl -s $BASE/audio_message/meditations -H "Authorization: Bearer $TOKEN"
# Expected: success:true, array

# Create series
SER=$(curl -s -X POST $BASE/audio_series \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Desk Check Series","author":"Pastor","description":"Test series","image_url":"https://example.com/series.jpg"}')
SER_ID=$(echo "$SER" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "SER_ID=$SER_ID"
# Expected: success:true

# Homepage (series + meditations)
curl -s $BASE/audio_series/home -H "Authorization: Bearer $TOKEN"
# Expected: success:true, {series:[...], meditations:[...]}

# Cleanup
curl -s -X DELETE $BASE/audio_message/delete/$MSG_ID -H "Authorization: Bearer $TOKEN"
curl -s -X DELETE $BASE/audio_message/meditation/delete/$MED_ID -H "Authorization: Bearer $TOKEN"
curl -s -X DELETE $BASE/audio_series/delete/$SER_ID -H "Authorization: Bearer $TOKEN"
```

---

## Subscription checks

```bash
# Create plan
PLAN=$(curl -s -X POST $BASE/subscription/plan \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Check Plan","type":0,"freq":3,"fee":5000,"currency":"NGN","code":"PLN_check01"}')
PLAN_ID=$(echo "$PLAN" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "PLAN_ID=$PLAN_ID"

# Create offering
OFF=$(curl -s -X POST $BASE/subscription/offering \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Check Offering"}')
OFF_ID=$(echo "$OFF" | grep -o '"ID":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "OFF_ID=$OFF_ID"

# Create plan offering
curl -s -X POST $BASE/subscription/plan/offering \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"subscription_plan_id\":\"$PLAN_ID\",\"subscription_offering_id\":\"$OFF_ID\",\"name\":\"Check PO\",\"type\":1,\"freq\":3,\"fee\":5000,\"currency\":\"NGN\",\"code\":\"PLN_check01\"}"
# Expected: success:true

curl -s $BASE/subscription/plan -H "Authorization: Bearer $TOKEN"       # list plans
curl -s $BASE/subscription/plan/$PLAN_ID -H "Authorization: Bearer $TOKEN"  # get plan
curl -s $BASE/subscription/plan/offering -H "Authorization: Bearer $TOKEN"   # list plan offerings
curl -s $BASE/subscription/offering -H "Authorization: Bearer $TOKEN"         # list offerings
curl -s $BASE/subscription -H "Authorization: Bearer $TOKEN"                  # list subscriptions

# Webhook — no sig (should 200 silently)
curl -s -o /dev/null -w "%{http_code}" -X POST $BASE/subscription/webhook \
  -H "Content-Type: application/json" \
  -d '{"event":"charge.success","data":{}}'
# Expected: 200

# Transaction (reaches Paystack — expect provider error in dev)
curl -s -X POST $BASE/subscription/transaction \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"plan_id\":\"$PLAN_ID\",\"email\":\"admin@example.com\"}"
# Expected: success:false (no live Paystack key) — correct in dev

# Admin global parameters
curl -s $BASE/admin/global -H "Authorization: Bearer $TOKEN"
curl -s -X PUT $BASE/admin/global \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"activate_subscription":true}'
# Expected: success:true both times

# Cleanup
curl -s -X DELETE $BASE/subscription/plan/$PLAN_ID -H "Authorization: Bearer $TOKEN"
curl -s -X DELETE $BASE/subscription/offering/delete/$OFF_ID -H "Authorization: Bearer $TOKEN"
```

---

## Score

After running, all endpoints should return:
- `"success":true` unless the test intentionally sends invalid data
- 200/201 for success cases
- 400/404 for expected validation/not-found errors
- 200 for Paystack webhook (always)
- Provider errors (500) only for Paystack transaction/verify in dev (no live key)

Any unexpected 500 is a bug — check server logs immediately.
