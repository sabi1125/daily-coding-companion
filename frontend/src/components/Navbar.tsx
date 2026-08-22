import { NavLink } from "react-router-dom"

const linkClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'text-foreground' : 'text-text-faint hover:text-foreground transition-colors'

function Navbar() {
  return (
    <nav className="flex items-center justify-between border border-border-faint px-10 py-4.5 sticky top-0 z-40 bg-background">
      <div className="text-base font-semibold">Coding Companion</div>
      <ul className="flex gap-6 text-sm font-medium">
        <li>
          <NavLink to="/" className={linkClassName}>Today</NavLink>
        </li>
        <li>
          <NavLink to="/history" className={linkClassName}>History</NavLink>
        </li>
        <li>
          <NavLink to="/settings" className={linkClassName}>Settings</NavLink>
        </li>
      </ul>
    </nav>
  )
}

export default Navbar
