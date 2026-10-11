import { Link, NavLink } from 'react-router-dom';

export function Header() {
  return (
    <header className="header">
      <div className="container header-content">
        <Link to="/store" className="logo" aria-label="Faultline Live Home">
          Faultline Live
        </Link>
        <nav className="nav" aria-label="Main navigation">
          <NavLink
            to="/store"
            className={({ isActive }) => `nav-link ${isActive ? 'active' : ''}`}
            end
          >
            Storefront
          </NavLink>
          <NavLink
            to="/ops"
            className={({ isActive }) => `nav-link ${isActive ? 'active' : ''}`}
          >
            Dashboard
          </NavLink>
        </nav>
      </div>
    </header>
  );
}