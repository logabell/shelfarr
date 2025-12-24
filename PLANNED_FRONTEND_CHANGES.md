# Frontend Integration Plan: Multi-API Strategy

## 1. Current State Analysis
The current frontend is tightly coupled to Hardcover.app as the single source of truth.
- **Search**: Exclusively uses `searchHardcover*` endpoints. Users cannot find books that aren't indexed in Hardcover.
- **Data Display**: `AuthorDetailPage` and `HardcoverAuthorPage` rely on Hardcover's metadata, including "physical only" flags which are often inaccurate for ebooks.
- **Adding Books**: The `addBook` function and API endpoint strictly require a `hardcoverId`, making it impossible to add books found solely via Open Library or Google Books.
- **Ebook Detection**: Relies on Hardcover's data, which lacks the comprehensive ebook availability data that Google Books provides.

## 2. Issues Found
1.  **False Negative "Physical Only"**: `HardcoverAuthorPage.tsx` displays a "Books have physical editions only" warning based solely on Hardcover data (line 391). This is causing user confusion (e.g., Graham Hancock case).
2.  **Search Blind Spots**: `SearchPage.tsx` defaults to Hardcover. If a book isn't on Hardcover, the user sees "No results", even if Open Library has it.
3.  **Inaccurate Status Badges**: Ebook/Audiobook badges in `BookDetailPage.tsx` are static and based on initial import data, not real-time availability from Google Books.
4.  **Add Flow Restriction**: The `BookResultCard` assumes every result has a `hardcoverId`. Open Library results (when integrated) will break this flow.

## 3. Recommended Changes

### A. Immediate Fixes (Low Risk)
1.  **Disable "Physical Only" Warning**: 
    - **File**: `frontend/src/pages/HardcoverAuthorPage.tsx`
    - **Action**: Comment out or remove the `physicalOnlyCount` check (lines 386-413).
    - **Why**: It's better to show nothing than false information.
2.  **Enable Open Library Search**:
    - **File**: `frontend/src/pages/SearchPage.tsx`
    - **Action**: Add a "Source" toggle or "Open Library" tab that calls `searchOpenLibrary`.
    - **Note**: "Add" button for these results must be disabled or routed to a "Find in Hardcover" helper until backend supports OL ID.

### B. Strategic Changes (High Impact)
1.  **Composite Book Card**:
    - **Component**: Create `CompositeBookCard.tsx`
    - **Logic**: 
        - Primary Metadata: Open Library (Title, Cover, Author)
        - Availability Badges: Google Books (`checkEbookStatus`) + Hardcover (Audio)
2.  **Real-time Availability Check**:
    - **File**: `frontend/src/pages/BookDetailPage.tsx`
    - **Action**: On mount, call `checkEbookStatus(isbn)` from Google Books.
    - **UI**: Show a verified "Google Books: Ebook Available" badge.

## 4. UI/UX Mockup Ideas

### The "Universal" Search Result
Instead of separate tabs, a unified list where:
- **Open Library** provides the main entry.
- **Badges** appear dynamically:
  - `[G] Ebook` (Green, validated via Google Books)
  - `[H] Series` (Blue, validated via Hardcover)
  - `[H] Audio` (Purple, validated via Hardcover)

### Author Page "Availability Scan"
A button on the Author Page: "Scan for Digital Editions".
- **Action**: Batches ISBNs on the page and checks Google Books.
- **Result**: Updates badges on the cards from "Unknown" to "Ebook Available".

## 5. Implementation Priority
1.  **Critical**: Remove "Physical Only" banner from `HardcoverAuthorPage.tsx`.
2.  **High**: Implement `searchOpenLibrary` in `SearchPage.tsx` (view-only mode).
3.  **Medium**: Add `checkEbookStatus` integration to `BookDetailPage.tsx`.
4.  **Future**: Update `addBook` flow to handle Open Library IDs (requires backend work).
