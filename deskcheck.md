# HOF Backend — Desk Check

All curl commands assume the server is running on `http://localhost:8080`.

```bash
BASE=http://localhost:8080
TOKEN=<paste token from sign_in response>
REFRESH=<paste refresh_token from sign_in response>
```

---

## Auth / Session (public)

```bash
# Sign up
curl -X POST $BASE/session/sign_up \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Test","last_name":"User","email":"user@example.com","password":"pass123"}'

# Sign in
curl -X POST $BASE/session/sign_in \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"pass123"}'

# Admin sign in
curl -X POST $BASE/session/sign_in/admin \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"admin123"}'

# Refresh token
curl -X POST $BASE/session/authenticate \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"

# Forgot password (sends OTP email)
curl -X POST $BASE/session/forgot_password \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'

# Verify OTP (from email)
curl -X PUT $BASE/session/verify_token \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","token":"123456"}'
```

---

## User (protected)

All requests require `-H "Authorization: Bearer $TOKEN"`.

```bash
# Get roles
curl $BASE/user/roles \
  -H "Authorization: Bearer $TOKEN"

# Assign roles
curl -X POST $BASE/user/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"roles":["steward","member"]}'

# Update profile
curl -X POST $BASE/user/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Updated","last_name":"Name"}'

# Reset password (no old password required)
curl -X POST $BASE/user/reset_password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"newpass123","password_confirm":"newpass123"}'

# Change password (requires old password)
curl -X POST $BASE/user/change_password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","old_password":"pass123","new_password":"newpass123","confirm_new_password":"newpass123"}'

# Send email verification link
curl -X POST $BASE/user/verify_email \
  -H "Authorization: Bearer $TOKEN"
```

### Favourites

```bash
# Add to favourites
curl -X POST $BASE/user/favourite \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"message_id":"<uuid>","series_id":"<uuid>","fav":true}'

# List favourites
curl $BASE/user/favourite/favs \
  -H "Authorization: Bearer $TOKEN"

# Remove from favourites
curl -X DELETE $BASE/user/favourite/delete/<message_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Devices

```bash
# Register device
curl -X POST $BASE/user/devices/add \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":"dev-001","who":"me","identifier":"abc123","os":"Android","brand":"Samsung","version":"13","status":"ACTIVE"}'

# List devices
curl $BASE/user/devices/all \
  -H "Authorization: Bearer $TOKEN"

# Update device status
curl -X PUT $BASE/user/devices/update/abc123/INACTIVE \
  -H "Authorization: Bearer $TOKEN"

# Delete device
curl -X DELETE $BASE/user/devices/delete/abc123 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Audio Messages (protected)

```bash
# Create
curl -X POST $BASE/audio_message \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Sunday Sermon","author":"Pastor John","description":"A great sermon","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":true}'

# List (supports ?search=, ?series_id=, ?page=, ?page_size=)
curl "$BASE/audio_message/?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# Get one
curl $BASE/audio_message/id/message/<message_id> \
  -H "Authorization: Bearer $TOKEN"

# Update
curl -X PUT $BASE/audio_message/update/<message_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Sermon","author":"Pastor John","description":"Updated","image_url":"https://example.com/img.jpg","audio_url":"https://example.com/audio.mp3","is_free":false}'

# Delete
curl -X DELETE $BASE/audio_message/delete/<message_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Meditations

```bash
# Create
curl -X POST $BASE/audio_message/meditation \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Morning Meditation","image":"https://example.com/med.jpg","status":"active"}'

# List
curl $BASE/audio_message/meditations \
  -H "Authorization: Bearer $TOKEN"

# Get one
curl $BASE/audio_message/meditation/<meditation_id> \
  -H "Authorization: Bearer $TOKEN"

# Update
curl -X PUT $BASE/audio_message/meditation/<meditation_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Evening Meditation","image":"https://example.com/med.jpg","status":"active"}'

# Delete
curl -X DELETE $BASE/audio_message/meditation/delete/<meditation_id> \
  -H "Authorization: Bearer $TOKEN"
```

---

## Audio Series (protected)

```bash
# Create
curl -X POST $BASE/audio_series \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Faith Series","author":"Pastor John","description":"A faith series","image_url":"https://example.com/series.jpg"}'

# List
curl "$BASE/audio_series/?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# Get one
curl $BASE/audio_series/id/series/<series_id> \
  -H "Authorization: Bearer $TOKEN"

# Homepage (series + meditations combined)
curl $BASE/audio_series/home \
  -H "Authorization: Bearer $TOKEN"

# Update
curl -X PUT $BASE/audio_series/update/<series_id> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Series","author":"Pastor John","description":"Updated","image_url":"https://example.com/series.jpg"}'

# Delete
curl -X DELETE $BASE/audio_series/delete/<series_id> \
  -H "Authorization: Bearer $TOKEN"
```

---

## Subscriptions (protected)

### Plans

```bash
# Create plan (type: 0=regular 1=premium; freq: 3=monthly 5=yearly)
curl -X POST $BASE/subscription/plan \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Monthly Plan","type":0,"freq":3,"fee":5000,"currency":"NGN","code":"PLN_xxx"}'

# List plans
curl $BASE/subscription/plan \
  -H "Authorization: Bearer $TOKEN"

# Get plan
curl $BASE/subscription/plan/<plan_id> \
  -H "Authorization: Bearer $TOKEN"

# Create plan offering
curl -X POST $BASE/subscription/plan/offering \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subscription_plan_id":"<plan_uuid>","subscription_offering_id":"<offering_uuid>","name":"Monthly Premium","type":1,"freq":3,"fee":5000,"currency":"NGN","code":"PLN_xxx"}'

# List plan offerings
curl $BASE/subscription/plan/offering \
  -H "Authorization: Bearer $TOKEN"

# Delete plan
curl -X DELETE $BASE/subscription/plan/<plan_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Offerings

```bash
# Create offering
curl -X POST $BASE/subscription/offering \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Premium Access"}'

# List offerings
curl $BASE/subscription/offering \
  -H "Authorization: Bearer $TOKEN"

# Delete offering
curl -X DELETE $BASE/subscription/offering/delete/<offering_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Subscriptions

```bash
# List all subscriptions (admin)
curl $BASE/subscription \
  -H "Authorization: Bearer $TOKEN"

# Initialize Paystack transaction (requires live Paystack credentials)
curl -X POST $BASE/subscription/transaction \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"plan_id":"<uuid>","email":"user@example.com"}'

# Verify subscription after payment (provide Paystack reference)
curl -X POST $BASE/subscription/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reference":"pay_ref_xxx"}'

# Disable subscription
curl -X DELETE $BASE/subscription/disable/<subscription_code> \
  -H "Authorization: Bearer $TOKEN"

# Paystack webhook (public — called by Paystack, not manually)
# POST $BASE/subscription/webhook  (verified by X-Paystack-Signature HMAC-SHA512)
```

---

## Admin (protected)

```bash
# Get global parameters
curl $BASE/admin/global \
  -H "Authorization: Bearer $TOKEN"

# Update global parameters
curl -X PUT $BASE/admin/global \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"activate_subscription":true}'
```

---

## File Upload (protected)

```bash
# Upload a file (returns S3 URL)
curl -X POST $BASE/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/file.mp3"
```

---

## Docs & Health

```bash
# Health check
curl $BASE/health

# Scalar UI (open in browser)
# http://localhost:8080/docs

# Swagger UI (open in browser)
# http://localhost:8080/swagger/index.html

# Raw OpenAPI JSON (import into Postman)
curl $BASE/swagger/doc.json -o hof_api.json
```
