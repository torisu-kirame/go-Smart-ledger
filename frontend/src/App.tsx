import { Route, Routes } from 'react-router-dom'
import { Layout } from './components/Layout'
import { Dashboard } from './pages/Dashboard'
import { LedgerDetail } from './pages/LedgerDetail'
import { Ledgers } from './pages/Ledgers'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="ledgers" element={<Ledgers />} />
        <Route path="ledgers/:id" element={<LedgerDetail />} />
      </Route>
    </Routes>
  )
}
