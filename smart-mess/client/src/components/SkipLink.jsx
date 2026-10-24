import { Link } from "react-router-dom";

export default function SkipLink() {
  return (
    <Link
      to="/"
      className="skip-link"
      onClick={(e) => {
        e.preventDefault();
        document.getElementById("main")?.focus();
      }}
    >
      Skip to main content
    </Link>
  );
}
