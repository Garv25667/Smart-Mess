import { useEffect } from "react";
import { Routes, Route, Navigate, NavLink, useLocation } from "react-router-dom";
import { useAuth } from "./auth/AuthContext.jsx";
import ProtectedRoute from "./auth/ProtectedRoute.jsx";
import SkipLink from "./components/SkipLink.jsx";
import { PlateIcon } from "./components/icons.jsx";
import Login from "./pages/Login.jsx";
import VerifyCode from "./pages/VerifyCode.jsx";
import ChooseGroup from "./pages/ChooseGroup.jsx";
import StudentHome from "./pages/StudentHome.jsx";
import ReviewScreen from "./pages/ReviewScreen.jsx";
import Settings from "./pages/Settings.jsx";
import ManagerCrowd from "./pages/ManagerCrowd.jsx";
import ManagerMenu from "./pages/ManagerMenu.jsx";
import ManagerReviews from "./pages/ManagerReviews.jsx";

const NAV_LINKS = [
  { to: "/", label: "Menu and crowd" },
  { to: "/review", label: "Review a meal" },
  { to: "/settings", label: "Settings" },
];

export default function App() {
  const { user } = useAuth();

  // Display settings come from the server user record only.
  const cvdMode = user?.cvdMode || "off";
  const reducedMotion = Boolean(user?.reducedMotion);
  const textScale = user?.textScale || 1.0;
  const theme = user?.theme || "system";
  useEffect(() => {
    document.documentElement.setAttribute("data-cvd", cvdMode);
    document.documentElement.setAttribute("data-theme", theme);
    document.documentElement.setAttribute("data-motion", reducedMotion ? "reduced" : "auto");
    document.documentElement.style.fontSize = `${(textScale * 100).toFixed(0)}%`;
  }, [cvdMode, reducedMotion, textScale, theme]);

  return (
    <>
      <SkipLink />
      <TopNav />
      <main id="main" tabIndex={-1} className="mx-auto max-w-3xl px-3 pb-16 pt-4">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/verify" element={<VerifyCode />} />
          <Route path="/choose-group" element={<ProtectedRoute><ChooseGroup /></ProtectedRoute>} />
          <Route path="/" element={<HomeEntry />} />
          <Route path="/review" element={<ProtectedRoute><ReviewScreen /></ProtectedRoute>} />
          <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
          <Route path="/manager" element={<ManagerRoleEntry />} />
          <Route path="/manager/crowd" element={<ProtectedRoute role="mess_manager"><ManagerCrowd /></ProtectedRoute>} />
          <Route path="/manager/menu" element={<ProtectedRoute role="main_manager"><ManagerMenu /></ProtectedRoute>} />
          <Route path="/manager/reviews" element={<ProtectedRoute role="main_manager"><ManagerReviews /></ProtectedRoute>} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>
    </>
  );
}

function HomeEntry() {
  const { user, loading } = useAuth();
  if (loading) {
    return (
      <p role="status" className="p-4 font-bold">
        Loading…
      </p>
    );
  }
  if (!user) return <Navigate to="/login" replace />;
  if (user.role === "mess_manager") return <Navigate to="/manager/crowd" replace />;
  if (user.role === "main_manager") return <Navigate to="/manager/menu" replace />;
  if (!user.hostelGroup) return <Navigate to="/choose-group" replace />;
  return <StudentHome />;
}

function ManagerRoleEntry() {
  const { user, loading } = useAuth();
  if (loading || !user) return <Navigate to="/" replace />;
  return user.role === "main_manager" ? <Navigate to="/manager/menu" replace /> : <Navigate to="/manager/crowd" replace />;
}

function TopNav() {
  const { user, logout } = useAuth();
  const { pathname } = useLocation();
  const role = user?.role || "guest";
  if (pathname === "/login" || pathname === "/verify") return null;
  return (
    <nav aria-label="Main" className="app-nav border-b-2 border-[var(--border)] bg-[var(--surface-muted)]">
      <div className="app-nav-inner mx-auto flex max-w-3xl flex-wrap items-center gap-2 px-3 py-1 text-sm">
        <span className="brand-mark" aria-hidden="true"><PlateIcon /></span>
        <span className="brand-name">Smart Mess</span>
        {role === "main_manager" ? (
          <>
            <NavLink to="/manager/menu" className="nav-link">Menu editor</NavLink>
            <NavLink to="/manager/reviews" className="nav-link">Reviews</NavLink>
          </>
        ) : role === "student" ? (
          NAV_LINKS.map((l) => (
            <NavLink key={l.to} to={l.to} end className="nav-link">
              {l.label}
            </NavLink>
          ))
        ) : role === "mess_manager" ? (
          <NavLink to="/manager/crowd" className="nav-link">Crowd</NavLink>
        ) : null}
        {user ? (
          <button type="button" onClick={logout} className="nav-link ml-auto">
            Log out
          </button>
        ) : null}
      </div>
    </nav>
  );
}
