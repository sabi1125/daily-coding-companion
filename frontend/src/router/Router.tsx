import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Today from '../features/today/Today.tsx'
import History from '../features/history/History.tsx'
import Settings from '../features/settings/Settings.tsx'
import Layout from './Layout.tsx'

function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Today />} />
          <Route path="/history" element={<History />} />
          <Route path="/settings" element={<Settings />} />
        </Route>
      </Routes >
    </BrowserRouter >
  )
}

export default Router
