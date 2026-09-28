// Copyright 2026 Phillip Cloud
// Licensed under the Apache License, Version 2.0

package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWarrantyStyleDateOnlyComparison(t *testing.T) {
	t.Parallel()
	// Warranty expires "2026-02-20". A user in UTC-5 at 23:00 local on
	// Feb 20 is still on the expiry date, so the warranty should be active.
	// But the absolute instant is Feb 21 04:00 UTC, which is After
	// midnight UTC Feb 20 -- the old code incorrectly shows expired.
	loc := time.FixedZone("UTC-5", -5*3600)
	now := time.Date(2026, 2, 20, 23, 0, 0, 0, loc) // Feb 20 23:00 local

	style := warrantyStyleAt("2026-02-20", now)
	assert.Equal(t, warrantyActive, style,
		"warranty expiring today should be active, not expired")
}

func TestWarrantyStyleExpiredNextDay(t *testing.T) {
	t.Parallel()
	// Same warranty, but now it's Feb 21 local -- should be expired.
	loc := time.FixedZone("UTC-5", -5*3600)
	now := time.Date(2026, 2, 21, 1, 0, 0, 0, loc)

	style := warrantyStyleAt("2026-02-20", now)
	assert.Equal(t, warrantyExpired, style,
		"warranty should be expired the day after expiry")
}

func TestUrgencyStyleUsesProvidedNow(t *testing.T) {
	t.Parallel()
	// Verify urgencyStyleAt uses the provided now, not real time.Now().
	// Set now to the day after the target -- should be overdue regardless
	// of what the real clock says.
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	style := urgencyStyleAt("2026-06-14", now)
	assert.Equal(t, urgencyOverdue, style,
		"item due yesterday (per provided now) should be overdue")
}

func TestUrgencyStyleDateOnlyComparison(t *testing.T) {
	t.Parallel()
	// A maintenance item due "2026-02-20". User in UTC-5 at 23:00 local
	// on Feb 19 -- locally that's 1 day away. But the absolute instant
	// is Feb 20 04:00 UTC, past midnight of the due date. The function
	// should compare local dates, not absolute instants.
	loc := time.FixedZone("UTC-5", -5*3600)
	now := time.Date(2026, 2, 19, 23, 0, 0, 0, loc) // Feb 19 23:00 local

	style := urgencyStyleAt("2026-02-20", now)
	assert.Equal(t, urgencySoon, style,
		"item due tomorrow (locally) should be 'soon', not overdue")
}

func TestUrgencyStyleThresholds_Overdue(t *testing.T) {
	t.Parallel()
	// Items due in the past (< 0 days) should be overdue (red bold).
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	style := urgencyStyleAt("2026-06-14", now) // -1 day
	assert.Equal(t, urgencyOverdue, style,
		"overdue item (-1 day) should use urgencyOverdue style")

	style = urgencyStyleAt("2026-06-01", now)   // -45 days
	assert.Equal(t, urgencyOverdue, style,
		"long overdue item (-45 days) should still be overdue")
}

func TestUrgencyStyleThresholds_Urgent(t *testing.T) {
	t.Parallel()
	// Items due in < 7 days (0-6 days) should be urgent (orange).
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	style := urgencyStyleAt("2026-06-15", now) // 0 days - today
	assert.Equal(t, urgencySoon, style,
		"item due today (0 days) should be urgent")

	style = urgencyStyleAt("2026-06-16", now) // +1 day
	assert.Equal(t, urgencySoon, style,
		"item due in 1 day should be urgent")

	style = urgencyStyleAt("2026-06-21", now) // +6 days (last day of urgent range)
	assert.Equal(t, urgencySoon, style,
		"item due in 6 days (< 7) should be urgent")
}

func TestUrgencyStyleThresholds_Upcoming(t *testing.T) {
	t.Parallel()
	// Items due in 7-15 days (7-14 days inclusive) should be upcoming (yellow).
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	style := urgencyStyleAt("2026-06-22", now) // +7 days
	assert.Equal(t, urgencyUpcoming, style,
		"item due in 7 days should be upcoming")

	style = urgencyStyleAt("2026-06-29", now) // +14 days (last day of upcoming range)
	assert.Equal(t, urgencyUpcoming, style,
		"item due in 14 days (< 15) should be upcoming")
}

func TestUrgencyStyleThresholds_Far(t *testing.T) {
	t.Parallel()
	// Items due >= 15 days should be far (green).
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)

	style := urgencyStyleAt("2026-06-30", now) // +15 days
	assert.Equal(t, urgencyFar, style,
		"item due in 15 days (>= 15) should be far")

	style = urgencyStyleAt("2026-07-15", now) // +30 days
	assert.Equal(t, urgencyFar, style,
		"item due in 30 days should be far")

	style = urgencyStyleAt("2026-12-31", now) // +541 days
	assert.Equal(t, urgencyFar, style,
		"item due far in the future should be far")
}
