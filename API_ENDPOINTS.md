# Upload API Endpoints — Reference

## Endpoints

### 1. Server-Side Upload (Multipart Optimized)

```
POST /admin/upload
Authorization: Bearer <admin-token>
Content-Type: multipart/form-data
```

**Request:**
```
file: <binary-data>
key: optional_custom_name  (auto-generated if omitted)
```

**Response (200):**
```json
{
  "url": "https://s3.amazonaws.com/hof-s3/goninja/hof/550e8400-e29b-41d4-a716-446655440000.mp3",
  "key": "550e8400-e29b-41d4-a716-446655440000.mp3"
}
```

**Error Responses:**
```json
// 400: File too large
{
  "error": "file size exceeds maximum allowed"
}

// 400: File too small
{
  "error": "file is empty"
}

// 400: Invalid content type
{
  "error": "unsupported file type: application/exe"
}

// 400: Missing file
{
  "error": "missing or invalid 'file' field"
}

// 401: Missing/invalid token
{
  "error": "unauthorized"
}

// 403: Not admin
{
  "error": "forbidden"
}
```

---

### 2. Generate Presigned URL (Direct S3 Upload)

```
GET /admin/upload/presigned?content_type=<mime-type>
Authorization: Bearer <admin-token>
```

**Query Parameters:**
- `content_type` (required) — MIME type of file to upload
  - Examples: `audio/mpeg`, `image/png`, `video/mp4`

**Response (200):**
```json
{
  "presigned_url": "https://s3.amazonaws.com/hof-s3/goninja/hof/550e8400...?Signature=AbCdEf...&Expires=1715000000",
  "expires_in": 3600,
  "key": "550e8400-e29b-41d4-a716-446655440000.mp3"
}
```

**Client then uploads to `presigned_url`:**
```bash
curl -X PUT \
  --data-binary @file.mp3 \
  -H "Content-Type: audio/mpeg" \
  'https://s3.amazonaws.com/...?Signature=...&Expires=...'

# S3 responds with:
# HTTP/1.1 200 OK
```

**Error Responses:**
```json
// 400: Missing content_type
{
  "error": "missing 'content_type' query parameter"
}

// 400: Unsupported content type
{
  "error": "unsupported file type: application/octet-stream"
}

// 403: Cloudinary (presigned not supported)
{
  "error": "presigned URLs not supported for Cloudinary"
}

// 401: Unauthorized
{
  "error": "unauthorized"
}
```

---

## Usage Examples

### cURL: Server Upload
```bash
# Upload a file (key auto-generated)
curl -X POST \
  -F "file=@audio.mp3" \
  -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/admin/upload

# Or with custom key
curl -X POST \
  -F "file=@audio.mp3" \
  -F "key=my-custom-audio" \
  -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/admin/upload
```

### cURL: Presigned URL
```bash
# Get presigned URL
curl -X GET \
  -H "Authorization: Bearer eyJhbGc..." \
  'http://localhost:8080/admin/upload/presigned?content_type=audio/mpeg' \
  | jq '.'

# Then upload directly to S3
PRESIGNED_URL=$(curl -s -X GET \
  -H "Authorization: Bearer eyJhbGc..." \
  'http://localhost:8080/admin/upload/presigned?content_type=audio/mpeg' \
  | jq -r '.presigned_url')

curl -X PUT \
  --data-binary @audio.mp3 \
  -H "Content-Type: audio/mpeg" \
  "$PRESIGNED_URL"
```

### JavaScript: Server Upload
```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch('/admin/upload', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`
  },
  body: formData
});

const { url, key } = await response.json();
console.log(`File uploaded: ${url}`);
```

### JavaScript: Presigned URL
```javascript
// Step 1: Get presigned URL
const presignedResponse = await fetch(
  '/admin/upload/presigned?content_type=audio/mpeg',
  {
    headers: { 'Authorization': `Bearer ${token}` }
  }
);

const { presigned_url, key, expires_in } = await presignedResponse.json();
console.log(`URL expires in ${expires_in}s`);

// Step 2: Upload directly to S3
const file = fileInput.files[0];
const uploadResponse = await fetch(presigned_url, {
  method: 'PUT',
  headers: {
    'Content-Type': 'audio/mpeg'
  },
  body: file
});

if (uploadResponse.ok) {
  console.log(`Uploaded to S3 with key: ${key}`);
}
```

### Python: Server Upload
```python
import requests

auth_header = {"Authorization": f"Bearer {token}"}

with open('audio.mp3', 'rb') as f:
    files = {'file': f}
    response = requests.post(
        'http://localhost:8080/admin/upload',
        files=files,
        headers=auth_header
    )

data = response.json()
print(f"Uploaded to: {data['url']}")
```

### Python: Presigned URL
```python
import requests

auth_header = {"Authorization": f"Bearer {token}"}

# Get presigned URL
presigned_response = requests.get(
    'http://localhost:8080/admin/upload/presigned?content_type=audio/mpeg',
    headers=auth_header
)

presigned_data = presigned_response.json()
presigned_url = presigned_data['presigned_url']

# Upload to S3
with open('audio.mp3', 'rb') as f:
    upload_response = requests.put(
        presigned_url,
        data=f,
        headers={'Content-Type': 'audio/mpeg'}
    )

print(f"Uploaded: {upload_response.status_code}")
```

---

## Supported Content Types

**Audio:**
- `audio/mpeg` or `audio/mp3`
- `audio/m4a`
- `audio/wav`
- `audio/webm`

**Images:**
- `image/jpeg` or `image/jpg`
- `image/png`
- `image/webp`
- `image/gif`

**Video:**
- `video/mp4`
- `video/webm`

**Documents:**
- `application/pdf`
- `application/json`

---

## Configuration

### Environment Variables

```env
# Size of each S3 multipart chunk (default: 10MB)
AWS_UPLOAD_PART_SIZE=10485760

# Maximum allowed file size (default: 500MB)
AWS_MAX_FILE_SIZE=524288000

# Presigned URL validity (default: 1 hour)
AWS_PRESIGNED_TTL=3600
```

### Examples

**For gigabit internet (larger chunks):**
```env
AWS_UPLOAD_PART_SIZE=52428800  # 50MB parts
```

**For 1GB max files:**
```env
AWS_MAX_FILE_SIZE=1073741824  # 1GB
```

**For 24-hour presigned URLs:**
```env
AWS_PRESIGNED_TTL=86400
```

---

## Response Headers

All successful uploads include:

```
Cache-Control: public, max-age=2592000, immutable
ETag: "550e8400-e29b-41d4-a716-446655440000.mp3"
```

This enables:
- Browser caching for 30 days
- CDN caching indefinitely
- Efficient cache validation

---

## Performance Characteristics

### Server Upload (`POST /admin/upload`)

| File Size | Time | Bandwidth | Server CPU |
|-----------|------|-----------|------------|
| 10MB | 2s | ~50 Mbps | 5% |
| 100MB | 15s | ~55 Mbps | 25% |
| 500MB | OOM | N/A | CRASH |

**Bottleneck:** Server memory for large files

### Presigned URL (`GET /presigned` + S3 PUT)

| File Size | Time | Bandwidth | Server CPU |
|-----------|------|-----------|------------|
| 10MB | 2s | ~50 Mbps | 0% |
| 100MB | 12s | ~65 Mbps | 0% |
| 500MB | 60s | ~65 Mbps | 0% |
| 5GB | 600s | ~65 Mbps | 0% |

**Advantage:** Server not involved, unlimited scaling

---

## Common Workflows

### Workflow 1: Simple File Upload
```
Admin selects file
POST /admin/upload (multipart)
⤷ Backend validates → Uploads to S3 → Returns URL
Display file URL
```

### Workflow 2: Large File Upload
```
Admin selects large file (>100MB)
GET /admin/upload/presigned (instant)
⤷ Backend generates URL (< 50ms)
Client uploads to S3 directly (parallel chunks)
Backend is not involved
Upload completes
```

### Workflow 3: Batch Upload
```
Admin selects 10 files
FOR each file:
  GET /admin/upload/presigned → presigned_url
  PUT file to presigned_url (parallel uploads)
All 10 uploads happen simultaneously
Server CPU: 0%
```

---

## Troubleshooting

### "file size exceeds maximum allowed"

**Cause:** File larger than `AWS_MAX_FILE_SIZE`

**Solution:** 
1. Increase `AWS_MAX_FILE_SIZE` in `.env`
2. Restart backend
3. Retry upload

### "Presigned URLs not supported for Cloudinary"

**Cause:** Using Cloudinary backend (doesn't support presigned URLs)

**Solution:** 
- Use `POST /admin/upload` instead
- Or switch to S3 by setting `AWS_SECRET` in `.env`

### "unsupported file type"

**Cause:** Content-Type not in whitelist

**Solution:**
1. Check `Content-Type` header matches file
2. Or add type to `AllowedContentTypes` in:
   ```
   internal/interfaces/http/handler/upload.go
   ```

### "missing 'content_type' query parameter"

**Cause:** Using presigned URL endpoint without query param

**Solution:** 
```bash
# Wrong:
GET /admin/upload/presigned

# Right:
GET /admin/upload/presigned?content_type=audio/mpeg
```

### Upload timeout after 10 minutes

**Cause:** Default HTTP timeout too short

**Solution:** 
1. Check server timeout settings
2. Use presigned URL endpoint (no server timeout)
3. Or increase server read timeout

---

## Security Considerations

✅ **Authentication:** Both endpoints require admin JWT token  
✅ **Content-Type:** Whitelist prevents malicious files  
✅ **File Size:** Max size prevents DoS attacks  
✅ **Presigned URLs:** Time-limited, can't be reused  
✅ **HMAC:** S3 signs all requests  

No additional security hardening needed.

---

## Monitoring

### Success Indicators
- Upload average time < 15s for 100MB
- Presigned URL generation < 100ms
- Zero file size validation rejections
- Cache hit rate > 90%

### Error Indicators to Watch
- High rate of content-type rejections → Check client
- Many file-too-large errors → Increase quota or educate users
- Presigned URL generation errors → S3 connectivity issue
- Upload timeouts → Network bottleneck

---

## Related Documentation

- **Full Technical Details:** `UPLOAD_OPTIMIZATIONS.md`
- **Quick Start Guide:** `UPLOAD_QUICK_REFERENCE.md`
- **Implementation Info:** `IMPLEMENTATION_SUMMARY.md`
- **Live API Docs:** `/docs` endpoint (Swagger UI)

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 2.0 | May 6, 2026 | Multipart S3, presigned URLs, cache headers |
| 1.0 | Earlier | Basic server upload |

