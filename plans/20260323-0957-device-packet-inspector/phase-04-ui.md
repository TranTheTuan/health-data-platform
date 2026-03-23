# Phase 4: Frontend UI

## Overview
- **Priority:** P1
- **Status:** Pending
- **Goal:** Create a clean, responsive page (`packets.html`) specifically designed for viewing, filtering, and paginating a device's raw telemetry. Maintain the system's current aesthetic and CSR loading logic.

## Core Requirements
1. Render a clean table interface showing: `Command Code`, `Raw Payload`, and `Timestamp`. 
   - A dropdown for 13 defined IW packet types + "All".
   - Two input fields for `Start Time` / `End Time` filtering.
2. Render manual paginators (`Items per page: 10/20/50/100`) and Prev/Next (`Offset: 0, 10, ...`) tracking `total_count`.
3. Link the feature directly from the main `dashboard.html` view.

## Related Code Files
- [web/templates/packets.html](file:///home/tuantt/projects/health-data-platform/web/templates/packets.html) (New)
- [web/templates/dashboard.html](file:///home/tuantt/projects/health-data-platform/web/templates/dashboard.html) (Modified link logic)

## Implementation Steps
1. Create `web/templates/packets.html` extending from `base.html` just like `dashboard.html`.
2. UI Layout Requirements:
   - Contains a Top Bar with an "⬅ Back to Dashboard" button.
   - Filter bar with `<select id="filter-type">` populated with all 13 command codes via static HTML `<option>` values:
     ```html
     <option value="">All Types</option>
     <option value="AP00">Login</option>
     <option value="AP01">GPS Location</option>
     <option value="APHT">Heart Rate + Blood Pressure</option>
     <!-- and the rest -->
     ```
   - Filter bar with `Start` & `End` HTML5 date/time inputs (`type="datetime-local"`).
   - "Refresh" button (`offset = 0`, trigger `fetchPackets()`).
3. JavaScript flow (`fetchPackets()`):
   - Grabs query variables from DOM inputs + `current_offset` + `current_limit`.
   - Disables table / shows "Loading..." spinner.
   - `fetch("/protected/devices/" + deviceId + "/packets?...")` array.
   - Updates `<tbody id="packet-list">` using row mapping.
   - Updates paginators (e.g. `Showing 11-20 of 500`).
4. Modify `dashboard.html`:
   - Within the table row builder (`renderDevices()`), update the column or action button to point to `/dashboard/devices/${dev.id}/packets`.

## Success Criteria
- Time parameters (if selected) send proper ISO string params to backend.
- Modifying "Page Size" defaults offset to 0.
- Clicking "Next Page" adds Limit to Offset.
- A user can easily read the raw payload of 10-100 items gracefully.
