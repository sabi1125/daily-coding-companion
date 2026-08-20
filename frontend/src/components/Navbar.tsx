import { NavLink } from "react-router-dom"

const linkClassName = ({ isActive }: { isActive: boolean }) =>
  isActive ? 'text-foreground' : 'text-muted-foreground'

function Navbar() {
  return (
    <nav className="flex items-center justify-between border-b border-border px-6 py-4">
      <div className="font-semibold">Coding Companion</div>
      <div>
        <ul className="flex gap-6">
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
      </div>
    </nav>
  )
}

export default Navbar
