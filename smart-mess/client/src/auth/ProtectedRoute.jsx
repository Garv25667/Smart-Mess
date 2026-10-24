// Route guard. Uses the server user (via AuthContext): role and hostel group
// decide where the user may be, mirroring the server's scoping rules.

import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "./AuthContext.jsx";

export default function ProtectedRoute({ children, role }) {
  const { user, loading } = useAuth();
  const location = useLocation();
  if (loading) {
    return (
      <p role="status" className="p-4 font-bold">
        Loading…
      </p>
    );
  }
  if (!user) return <Navigate to="/login" replace />;
  // Redirect to the group chooser unless we are already there (a same-path
  // Navigate would loop and render nothing else, blanking the page).
  if (
    user.role === "student" &&
    !user.hostelGroup &&
    location.pathname !== "/choose-group"
  ) {
    return <Navigate to="/choose-group" replace />;
  }
  // Mirror the server rule: only a main manager may open manager-only routes.
  if (role && user.role !== role) return <Navigate to="/" replace />;
  return children;
}
