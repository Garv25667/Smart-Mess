// Real API client (ARCHITECTURE.md section 7). Every call hits the Go backend;
// auth is the httpOnly session cookie (credentials: "include").
// Errors always carry the server's error code and message.

export const API_URL =
  (typeof import.meta !== "undefined" && import.meta.env && import.meta.env.VITE_API_URL) ||
  "http://localhost:8080/api";

export class ApiError extends Error {
  constructor(code, message, status) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

async function api(path, { method = "GET", body } = {}) {
  const res = await fetch(`${API_URL}${path}`, {
    method,
    credentials: "include",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (res.status === 204) return {};
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = data && data.error ? data.error : {};
    throw new ApiError(err.code || "UNKNOWN", err.message || `Request failed (${res.status})`, res.status);
  }
  return data;
}

// ---------- auth ----------

export function requestOTP(email) {
  return api("/auth/request-otp", { method: "POST", body: { email } });
}

export function verifyOTP(email, code) {
  return api("/auth/verify-otp", { method: "POST", body: { email, code } });
}

export function logout() {
  return api("/auth/logout", { method: "POST" });
}

// ---------- current user ----------

export function getMe() {
  return api("/me");
}

export function updatePreferences(patch) {
  return api("/me/preferences", { method: "PATCH", body: patch });
}

export function setHostelGroup(hostelGroup) {
  return api("/me/hostel-group", { method: "POST", body: { hostelGroup } });
}

// ---------- student reads ----------

export function getMesses() {
  return api("/messes");
}

export function getMeals(date) {
  return api(`/meals${date ? `?date=${date}` : ""}`);
}

export function getMealReviews(mealId, variant) {
  const q = new URLSearchParams();
  if (variant) q.set("variant", variant);
  return api(`/meals/${mealId}/reviews?${q}`);
}

// ---------- reviews ----------

export function submitReview(mealId, { variant, stars, comment }) {
  return api(`/meals/${mealId}/reviews`, { method: "POST", body: { variant, stars, comment } });
}

// ---------- managers ----------

export function setCrowd(messId, level) {
  return api(`/manager/messes/${messId}/crowd`, { method: "PUT", body: { level } });
}

export function getManagerMeals(date) {
  return api(`/manager/meals${date ? `?date=${date}` : ""}`);
}

export function getDefaultItems(mealType) {
  return api(`/manager/default-items?mealType=${mealType}`);
}

export function saveMeal(date, mealType, items) {
  return api(`/manager/meals/${date}/${mealType}`, { method: "PUT", body: { items } });
}

export function publishMeal(date, mealType) {
  return api(`/manager/meals/${date}/${mealType}/publish`, { method: "POST" });
}

export function getManagerReviews(mealId) {
  return api(`/manager/reviews?mealId=${mealId}`);
}

export function moderateReview(reviewId, hide) {
  return api(`/manager/reviews/${reviewId}/${hide ? "hide" : "unhide"}`, { method: "POST" });
}

// ---------- helpers shared by screens ----------

// "12:30" from an ISO instant. The server sends Asia/Kolkata wall clock inside
// the string (…T12:30:00+05:30), so read it from the string, never the device
// clock or the browser timezone (server is the only source of "now").
export function timeHM(iso) {
  if (!iso) return "--:--";
  const m = /(\d{2}):(\d{2})(?:[+-]\d{2}:?\d{2}|Z)?$/.exec(String(iso));
  if (!m) return "--:--";
  return `${m[1]}:${m[2]}`;
}

export function scheduleLabel(meal) {
  return `${timeHM(meal.startsAt)}–${timeHM(meal.endsAt)}`;
}

// YYYY-MM-DD for "today" as the server sees it (the client's Kolkata browser
// clock is the display date; the server decides what is current).
export function todayStr() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}
