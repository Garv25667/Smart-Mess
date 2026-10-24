// Auth context: one GET /api/me on load, exposed to every screen. The role and
// hostel group always come from the server user record, never from the client.

import { createContext, useCallback, useContext, useEffect, useState } from "react";
import { getMe, logout as apiLogout, ApiError } from "../api";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    try {
      const data = await getMe();
      setUser(data.user);
      return data.user;
    } catch (err) {
      // 401 from /api/me (or any request failure) == no usable session:
      // the user is unauthenticated, the login page will show.
      setUser(null);
      if (!(err instanceof ApiError && err.status === 401)) console.error(err);
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const logout = useCallback(async () => {
    try {
      await apiLogout();
    } finally {
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider value={{ user, setUser, loading, refresh, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
