# 🔧 Logout Cookie Fix

## Problem

The logout button wasn't working in production - the authentication cookie remained in the browser after clicking logout.

## Root Cause

The `Logout` handler in `internal/handlers/authHandler.go` was setting a cookie to expire it, but it wasn't matching the attributes of the original cookie:

**Original Login/Register Cookie:**
- `Secure`: `true` (in production HTTPS)
- `SameSite`: `http.SameSiteNoneMode` (required for cross-origin)
- `HttpOnly`: `true`
- `Path`: `/`

**Broken Logout Cookie:**
- `Secure`: **missing** ❌
- `SameSite`: **missing** ❌
- `HttpOnly`: `true`
- `Path`: `/`
- `MaxAge`: `-1` (to expire)

**Issue:** In production (HTTPS), browsers require the `Secure` and `SameSite` attributes to match when clearing a cookie. Without these attributes, the browser treats it as a different cookie and the original remains.

## Solution

Updated the `Logout` handler to match all cookie attributes:

```go
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Determine if we're in production (HTTPS)
	isProduction := r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,          // Must match the original cookie
		SameSite: http.SameSiteNoneMode, // Must match the original cookie
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
```

## Verification

Checked all cookie-related operations in the codebase:

### ✅ All Cookie Operations Verified

1. **Register** (`authHandler.go` line 58)
   - ✅ `Secure: isProduction`
   - ✅ `SameSite: http.SameSiteNoneMode`
   - ✅ Correctly configured

2. **Login** (`authHandler.go` line 112)
   - ✅ `Secure: isProduction`
   - ✅ `SameSite: http.SameSiteNoneMode`
   - ✅ Correctly configured

3. **Logout** (`authHandler.go` line 138) - **FIXED**
   - ✅ Now includes `Secure: isProduction`
   - ✅ Now includes `SameSite: http.SameSiteNoneMode`
   - ✅ Matches login/register cookies

4. **Cookie Reading** (`authMiddleware.go` line 19)
   - ✅ Just reads the cookie (no issues)
   - ✅ No changes needed

## Testing Checklist

After deploying this fix, test the following:

- [ ] **Desktop Logout**: Click logout button on desktop view
- [ ] **Mobile Logout**: Click logout button in mobile menu
- [ ] **Verify Cookie Cleared**: Check browser DevTools → Application → Cookies
  - The `auth_token` cookie should be gone
- [ ] **Redirect to Login**: Should redirect to login page
- [ ] **No Access to Protected Routes**: Try accessing `/library` - should redirect to login
- [ ] **Can Login Again**: Login should work normally after logout

## Browser Testing

Test in multiple browsers to ensure compatibility:
- [ ] Chrome/Edge (Chromium)
- [ ] Firefox
- [ ] Safari
- [ ] Mobile Safari (iOS)
- [ ] Mobile Chrome (Android)

## Files Changed

1. `/internal/handlers/authHandler.go`
   - Modified `Logout` function to include `Secure` and `SameSite` attributes

## Additional Changes

Updated GitHub Actions workflows to trigger on `prod` branch instead of `main`:

1. `/.github/workflows/deploy-railway.yml` - Backend deployment
2. `/.github/workflows/deploy-frontend.yml` - Frontend deployment

This aligns with Railway CI configuration that watches the `prod` branch.

## Deployment

```bash
# Backend
cd booklib-backend
git checkout -b prod  # If prod branch doesn't exist
git add .
git commit -m "fix: logout cookie attributes for production HTTPS"
git push origin prod

# Frontend (if updating workflows)
cd booklib-frontend
git checkout -b prod  # If prod branch doesn't exist
git add .github/workflows/deploy-frontend.yml
git commit -m "chore: update deployment workflow to prod branch"
git push origin prod
```

Railway will automatically deploy when changes are pushed to `prod` branch.

## Notes

- This issue only affected production (HTTPS) environments
- Local development (HTTP) was working because `Secure` flag defaults to `false`
- The fix ensures consistent cookie attributes across login, register, and logout operations
- The `isProduction` check uses `X-Forwarded-Proto` header (set by Railway/reverse proxies) or `r.TLS`

## References

- [MDN: Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie)
- [MDN: SameSite cookies](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie/SameSite)
- [Chrome Cookie Behavior](https://web.dev/samesite-cookies-explained/)
