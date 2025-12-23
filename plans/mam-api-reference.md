# MyAnonaMouse (MAM) API Reference

## Torrent Search (JSON)

**URI:** `/tor/js/loadSearchJSONbasic.php`

Search torrents on MAM, returning JSON results. This endpoint accepts input via:
- POST as `application/xml`, `application/json`, `multipart/form-data`, or `application/x-www-form-urlencoded`
- GET parameters

---

## Input Parameters

| Parameter | Data Type | Description |
|-----------|-----------|-------------|
| `description` | empty | If set, displays the full description field for the torrent |
| `dlLink` | blank | Show hash for download link (prepend `https://www.myanonamouse.net/tor/download.php/` to use) |
| `isbn` | set | If set, returns the ISBN field (often blank) |
| `mediaInfo` | set | When set, returns key parts of mediaInfo |
| `my_snatched` | exists | If set, limits results only to what you have snatched |
| `perpage` | int | Range 5-1000, number of results to return |

### `tor` Array (REQUIRED)

| Parameter | Data Type | Description |
|-----------|-----------|-------------|
| `browse_lang` | list | List of integers for languages to view |
| `cat` | list | List of integers for categories to view |
| `endDate` | date | Format `YYYY-MM-DD` or unix timestamp, torrents created before (exclusive) |
| `hash` | HEX string | Hexadecimal encoded hash from a torrent |
| `id` | int | Return only a single ID's data |
| `main_cat` | array | Array of main category IDs: `13`=AudioBooks, `14`=E-Books, `15`=Musicology, `16`=Radio |
| `searchIn` | enum | (list to come) |
| `searchType` | enum | See Search Types below |
| `sortType` | enum | See Sort Types below |
| `startDate` | date | Format `YYYY-MM-DD` or unix timestamp, earliest torrents (inclusive) |
| `startNumber` | int | Number of entries to skip (pagination) |
| `text` | text | Text to search for |

### `srchIn` Array (Search Fields)

| Parameter | Description |
|-----------|-------------|
| `author` | Search in author field |
| `description` | Search in description field |
| `filenames` | Search in filenames |
| `fileTypes` | Search in file types |
| `narrator` | Search in narrator field |
| `series` | Search in series field |
| `tags` | Search in tags field |
| `title` | Search in title field |

### Search Types (`searchType`)

| Value | Description |
|-------|-------------|
| `all` | Search everything |
| `active` | Last update had 1+ seeders |
| `inactive` | Last update has 0 seeders |
| `fl` | Freeleech torrents |
| `fl-VIP` | Freeleech or VIP torrents |
| `VIP` | VIP torrents |
| `nVIP` | Torrents not VIP |
| `nMeta` | Torrents missing metadata (old torrents) |

### Sort Types (`sortType`)

| Value | Description |
|-------|-------------|
| `titleAsc` / `titleDesc` | By title |
| `fileAsc` / `fileDesc` | By number of files |
| `sizeAsc` / `sizeDesc` | By size |
| `seedersAsc` / `seedersDesc` | By number of seeders |
| `leechersAsc` / `leechersDesc` | By number of leechers |
| `snatchedAsc` / `snatchedDesc` | By times snatched |
| `dateAsc` / `dateDesc` | By date added |
| `bmkaAsc` / `bmkaDesc` | By date bookmarked |
| `reseedAsc` / `reseedDesc` | By reseed request date |
| `categoryAsc` / `categoryDesc` | By category number |
| `random` | Random order |
| `default` | Context-dependent default sorting |

---

## Example Input

### URL Encoded (GET/POST)
```
tor%5Bcat%5D%5B%5D=0&tor%5BsortType%5D=default&tor%5BbrowseStart%5D=true&tor%5BstartNumber%5D=0&bannerLink&bookmarks&dlLink&description&tor%5Btext%5D=mp3%20m4a
```

### JSON (POST)
```json
{
    "tor": {
        "text": "collection cookbooks food test kitchen",
        "srchIn": ["title", "author", "narrator"],
        "searchType": "all",
        "searchIn": "torrents",
        "cat": ["0"],
        "browseFlagsHideVsShow": "0",
        "startDate": "",
        "endDate": "",
        "hash": "",
        "sortType": "default",
        "startNumber": "0"
    },
    "thumbnail": "true"
}
```

---

## Output Parameters

### Response Structure
```json
{
    "data": [...],
    "total": 100,
    "total_found": 100
}
```

### Data Object Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | int | Torrent ID. Access page via `/t/{id}` |
| `title` | string | Torrent title |
| `name` | string | Alternative title field |
| `size` | string | Size of the torrent (in bytes as string) |
| `seeders` | int | Number of seeders |
| `leechers` | int | Number of leechers |
| `numfiles` | int | Number of files in torrent |
| `times_completed` | int | Number of completed snatches |
| `added` | datetime | Upload date/time (UTC) |
| `category` | int | Specific category ID |
| `catname` | string | Category name |
| `main_cat` | int | Main category: `13`=AudioBooks, `14`=E-Books, `15`=Musicology, `16`=Radio |
| `author_info` | string | JSON object of `id:name` pairs |
| `narrator_info` | string | JSON object of `id:name` pairs |
| `series_info` | string | JSON object of `id:[name, position]` pairs |
| `tags` | string | Space-separated tags |
| `filetype` | string | File types (e.g., "mp3 m4a") |
| `lang_code` | string | 3-letter ISO language code |
| `language` | int | Internal language ID |
| `vip` | boolean | If item is VIP |
| `free` | boolean | If item is freeleech |
| `fl_vip` | boolean | If item is freeleech and/or VIP |
| `dl` | string | User-specific hash for download. Use: `/tor/download.php/{dl}` |
| `description` | string | Full description (if requested) |
| `my_snatched` | boolean | Whether you've snatched this torrent |
| `personal_freeleech` | boolean | Whether you've bought personal freeleech |
| `bookmarked` | datetime/null | When bookmarked (or null) |
| `owner` | int | Uploader user ID |
| `owner_name` | string | Uploader username |
| `comments` | int | Number of comments |
| `browseflags` | bitfield | Tags and flags bitfield |

---

## Download URLs

Two methods to download torrents:

1. **With session cookie:** `https://www.myanonamouse.net/tor/download.php?tid={id}`
2. **Without session cookie:** `https://www.myanonamouse.net/tor/download.php/{dl}`

Where `{dl}` is the user-specific download hash from the search response.

---

## Example Response

```json
{
    "data": [
        {
            "id": "273200",
            "language": "1",
            "main_cat": "13",
            "category": "108",
            "catname": "Audiobooks - Urban Fantasy",
            "size": "6324306932",
            "numfiles": "149",
            "vip": "0",
            "free": "0",
            "fl_vip": "0",
            "name": "Love at Stake series",
            "tags": "Love at Stake series unabridged 64–128 Kbps Fiction...",
            "author_info": "{\"8234\": \"Kerrelyn Sparks\"}",
            "narrator_info": "{\"1\": \"Abby Craden\", ...}",
            "series_info": "{\"67\": [\"Love at Stake\", \"01-16, 13.5\"]}",
            "filetype": "m4a mp3",
            "dl": "k0S0Bxpdds1Q1vAvRI,ILhtA5UvR...",
            "bookmarked": null
        }
    ],
    "total": 100,
    "total_found": 100
}
```

---

## Important Notes

1. **Field types are inconsistent** - Some fields like `id`, `language`, `category`, `size`, `seeders`, etc. may be returned as either strings OR numbers depending on context. Always handle both types.

2. **Boolean fields** - Fields like `vip`, `free`, `fl_vip`, `my_snatched`, `personal_freeleech` may be returned as `"0"`/`"1"` strings, `0`/`1` integers, or actual booleans.

3. **Main Categories:**
   - `13` = AudioBooks
   - `14` = E-Books  
   - `15` = Musicology
   - `16` = Radio

4. **Authentication** - Requires valid session cookie (`mam_id`) for API access.
