# Book APIs Comparison: Limitations & Rate Limits

## Google Books API vs Hardcover API vs Open Library API

This document provides a comprehensive comparison of the limitations, rate limits, and restrictions for three major book APIs.

---

## Quick Comparison Table

| Feature | Google Books API | Hardcover API | Open Library API |
|---------|-----------------|---------------|------------------|
| **API Type** | REST | GraphQL (Hasura) | REST |
| **Authentication** | API Key / OAuth 2.0 | Bearer Token | None (read) / S3 Keys (write) |
| **Cost** | Free (with limits) | Free | Free |
| **Rate Limit** | 1,000/day default | 60/minute | Soft limits + User-Agent required |
| **Max Results/Request** | 40 | N/A (pagination) | 1,000 |
| **Status** | Production | Beta | Production |
| **Data Scope** | Public + User libraries | Personal + Public + Followed | Public catalog |

---

## Google Books API Limitations

### Rate Limits

| Limit Type | Value | Notes |
|------------|-------|-------|
| **Daily quota** | 1,000 requests/day | Default for API key |
| **Per-user limit** | 100 requests/100 seconds | Per authenticated user |
| **Max results per request** | 40 | Use pagination for more |
| **Quota reset** | Midnight Pacific Time | Daily reset |

### Data Limitations

| Limitation | Details |
|------------|---------|
| **Geographic restrictions** | Content availability varies by country |
| **Preview availability** | Not all books have previews |
| **No bulk queries** | Must query books individually or in small batches |
| **Metadata completeness** | Some books have incomplete data |
| **Cover images** | Not available for all books |
| **Full text access** | Limited to preview/snippet based on publisher settings |

### API Restrictions

| Restriction | Details |
|-------------|---------|
| **No commercial redistribution** | Cannot resell or redistribute data |
| **Attribution required** | Must display "Powered by Google" |
| **Terms of Service** | Must comply with Google API ToS |
| **Response caching** | Limited caching allowed per ToS |
| **User data access** | Requires OAuth 2.0 for user libraries |

### Search Limitations

| Limitation | Value |
|------------|-------|
| **startIndex maximum** | Limited (not officially documented, ~1000) |
| **Query length** | Standard URL length limits |
| **Special characters** | Must be URL-encoded |
| **Filter combinations** | Some filter combinations may not work |

### Error Scenarios

| Error | Cause |
|-------|-------|
| 403 Forbidden | Quota exceeded or API key invalid |
| 400 Bad Request | Invalid parameters |
| 404 Not Found | Book/volume not found |
| 429 Too Many Requests | Rate limit exceeded |

---

## Hardcover API Limitations

### Rate Limits

| Limit Type | Value | Notes |
|------------|-------|-------|
| **Requests per minute** | 60 | Hard limit |
| **Query timeout** | 30 seconds | Queries exceeding this will fail |
| **Query depth (2025)** | 3 levels max | Nested query limitation |

### Token Limitations

| Limitation | Details |
|------------|---------|
| **Token expiry** | 1 year (resets January 1st) |
| **Token resets** | May be reset without notice during beta |
| **Token sharing** | Strictly prohibited |
| **Token security** | Can access/delete entire account |

### Access Restrictions

| Restriction | Details |
|-------------|---------|
| **Data ownership** | Can only access/modify your own data |
| **Browser execution** | **NOT allowed** - security risk |
| **Allowed origins** | Only `localhost` or backend APIs |
| **User data (2025)** | Limited to own data, public data, and followed users |
| **Backend only** | Must run from server-side code |
| **Offline use only** | Currently restricted to offline/local use |

### Disabled Query Operators

The following operators are **permanently disabled**:

| Operator | Description |
|----------|-------------|
| `_like` | Pattern matching |
| `_nlike` | Negative pattern matching |
| `_ilike` | Case-insensitive pattern matching |
| `_nilike` | Negative case-insensitive pattern |
| `_regex` | Regular expression matching |
| `_nregex` | Negative regex matching |
| `_iregex` | Case-insensitive regex |
| `_niregex` | Negative case-insensitive regex |
| `_similar` | Similar pattern matching |
| `_nsimilar` | Negative similar pattern |

### API Status Warnings

| Warning | Details |
|---------|---------|
| **Beta status** | API is heavily in flux |
| **Breaking changes** | Features may break without notice |
| **No stability guarantee** | Schema and endpoints may change |
| **Limited documentation** | Documentation still being developed |

### Data Access Limits

| Limitation | Details |
|------------|---------|
| **Cross-user data** | Cannot access other users' private data |
| **Mutations** | Limited to own account data |
| **Search depth** | Limited query complexity |

---

## Open Library API Limitations

### Rate Limits

| Limit Type | Value | Notes |
|------------|-------|-------|
| **General API** | No strict limit | But User-Agent required for frequent use |
| **Covers API (by ISBN/OCLC/LCCN)** | 100 requests/5 minutes | Per IP address |
| **Covers API (by Cover ID/OLID)** | No limit | Use these for high volume |
| **Query results default** | 100 items | Per request |
| **Query results maximum** | 1,000 items | Hard limit |
| **Query offset maximum** | 10,000 | Cannot paginate beyond |
| **History entries maximum** | 1,000 | Per request |

### Required Headers

| Requirement | Details |
|-------------|---------|
| **User-Agent** | **REQUIRED** for frequent use (multiple calls/minute) |
| **Format** | Must include app name AND contact email/phone |
| **Consequence** | Apps without User-Agent may be blocked |

### Data Access Restrictions

| Restriction | Details |
|-------------|---------|
| **No bulk downloads via API** | Use monthly data dumps instead |
| **Write operations** | Require S3 key authentication |
| **Schema stability** | Search schema not guaranteed stable |
| **Data completeness** | Many books have incomplete metadata |

### Cover Image Restrictions

| Restriction | Details |
|-------------|---------|
| **Rate limit by identifier** | 100/5min for ISBN, OCLC, LCCN lookups |
| **No rate limit** | Cover ID and OLID lookups |
| **Bulk downloads prohibited** | Use archive.org cover dumps |
| **403 Forbidden** | Returned when rate limit exceeded |

### API Stability

| Concern | Details |
|---------|---------|
| **Schema changes** | Fields may be added/removed |
| **Downtime** | Service may have occasional outages |
| **Performance** | Response times can vary |
| **Data accuracy** | Community-edited, may contain errors |

### Write Operation Limits

| Operation | Requirement |
|-----------|-------------|
| **Adding books** | S3 key authentication |
| **Editing books** | S3 key authentication |
| **Creating lists** | Account required |
| **Internal API** | PUT/POST only from localhost |

---

## Detailed Feature Comparison

### Authentication Comparison

| Feature | Google Books | Hardcover | Open Library |
|---------|-------------|-----------|--------------|
| **Public data access** | API Key | Bearer Token | None required |
| **User data access** | OAuth 2.0 | Bearer Token | S3 Keys |
| **Token expiration** | Varies | 1 year | Never (S3) |
| **Browser-safe** | Yes | **No** | Yes (read only) |

### Search Capabilities

| Feature | Google Books | Hardcover | Open Library |
|---------|-------------|-----------|--------------|
| **Full-text search** | Yes | Yes (Typesense) | Yes (Solr) |
| **Field-specific search** | Yes | Yes | Yes |
| **Regex/pattern matching** | No | **Disabled** | No |
| **Fuzzy search** | Limited | Yes | Yes |
| **Max results** | 40/request | Pagination | 1,000/request |

### Data Availability

| Data Type | Google Books | Hardcover | Open Library |
|-----------|-------------|-----------|--------------|
| **Book metadata** | Extensive | Community-driven | Extensive |
| **Cover images** | Yes | Yes (cached) | Yes |
| **Author photos** | No | Yes | Yes |
| **User ratings** | Yes | Yes | Yes |
| **Reviews** | Yes (Google) | Yes | No |
| **Reading progress** | Yes | Yes | Yes |
| **Book previews** | Yes | No | Yes (Archive.org) |
| **Purchase links** | Yes | No | No |

### User Features

| Feature | Google Books | Hardcover | Open Library |
|---------|-------------|-----------|--------------|
| **Reading lists** | Yes | Yes | Yes |
| **Custom shelves** | Yes | Yes | Yes |
| **Reading progress** | Yes | Yes | Yes |
| **Reviews/ratings** | Yes | Yes | Yes |
| **Social features** | Limited | Yes | Limited |
| **Lists/collections** | Yes | Yes | Yes |

---

## Rate Limit Summary

### Requests Per Time Period

| API | Limit | Period | Notes |
|-----|-------|--------|-------|
| **Google Books** | 1,000 | Day | Default quota |
| **Google Books** | 100 | 100 seconds | Per user |
| **Hardcover** | 60 | Minute | All requests |
| **Open Library** | ~100 | 5 minutes | Covers by ISBN only |
| **Open Library** | Soft limit | Varies | General API |

### Results Per Request

| API | Maximum | Default |
|-----|---------|---------|
| **Google Books** | 40 | 10 |
| **Hardcover** | Unlimited (pagination) | Varies |
| **Open Library** | 1,000 | 100 |

### Query Timeouts

| API | Timeout |
|-----|---------|
| **Google Books** | Standard HTTP |
| **Hardcover** | 30 seconds |
| **Open Library** | Standard HTTP |

---

## Compliance Requirements

### Google Books API

- Must display "Powered by Google" attribution
- Cannot resell or redistribute data
- Must comply with Google API Terms of Service
- Cannot cache responses beyond allowed period
- Cannot use for competing book service

### Hardcover API

- Must keep token private
- Backend use only (no browser)
- Cannot access other users' data
- Must respect rate limits
- Data ownership remains with original user

### Open Library API

- Must include User-Agent header
- Cannot bulk download via API
- Must use data dumps for large imports
- Attribution to Open Library appreciated
- Community contribution encouraged

---

## Recommendations by Use Case

### High-Volume Applications

| Use Case | Recommended API | Reason |
|----------|-----------------|--------|
| **Book catalog** | Open Library | No strict limits, bulk dumps available |
| **User tracking app** | Hardcover | Best social/tracking features |
| **Commercial product** | Google Books | Most stable, best ToS for commercial |

### Low-Volume Applications

| Use Case | Recommended API | Reason |
|----------|-----------------|--------|
| **Personal project** | Any | All free for low volume |
| **Book lookup widget** | Google Books | Simple API, good coverage |
| **Reading tracker** | Hardcover | Purpose-built for tracking |

### Specific Features

| Need | Best API | Alternative |
|------|----------|-------------|
| **Book covers** | Open Library | Google Books |
| **Author photos** | Open Library | Hardcover |
| **Reading progress** | Hardcover | Google Books |
| **Book previews** | Google Books | Open Library (Archive.org) |
| **Social features** | Hardcover | None |
| **Purchase info** | Google Books | None |

---

## Error Code Comparison

| HTTP Code | Google Books | Hardcover | Open Library |
|-----------|-------------|-----------|--------------|
| **200** | Success | Success | Success |
| **400** | Bad Request | Bad Request | Bad Request |
| **401** | Unauthorized | Invalid Token | N/A |
| **403** | Quota/Forbidden | Access Denied | Rate Limited |
| **404** | Not Found | Not Found | Not Found |
| **429** | Rate Limited | Throttled | N/A |
| **500** | Server Error | Server Error | Server Error |

---

## Summary: Key Limitations to Remember

### Google Books API
1. ✗ 1,000 requests/day default limit
2. ✗ Maximum 40 results per request
3. ✗ Geographic content restrictions
4. ✗ Attribution required
5. ✓ Most stable and documented

### Hardcover API
1. ✗ 60 requests/minute hard limit
2. ✗ 30-second query timeout
3. ✗ **No browser execution allowed**
4. ✗ Beta status - may break
5. ✗ Token expires yearly
6. ✗ Regex/pattern operators disabled
7. ✗ Query depth limited to 3 levels (2025)

### Open Library API
1. ✗ User-Agent header **required** for frequent use
2. ✗ 100 cover requests/5 min (by ISBN)
3. ✗ 1,000 results maximum per query
4. ✗ No bulk downloads via API
5. ✗ Schema not guaranteed stable
6. ✓ Most permissive for high volume

---

*Document generated: December 2024*
