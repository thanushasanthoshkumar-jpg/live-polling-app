import React, { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import api, { getErrorMessage } from '../api.js'

export default function Dashboard() {
  const [polls, setPolls] = useState([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const res = await api.get('/polls')
        if (!cancelled) setPolls(res.data || [])
      } catch (err) {
        if (!cancelled) setError(getErrorMessage(err))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [])

  function copyLink(id) {
    const url = `${window.location.origin}/poll/${id}`
    navigator.clipboard?.writeText(url)
  }

  return (
    <div className="max-w-3xl mx-auto px-4 py-10">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Your polls</h2>
        <Link
          to="/create"
          className="px-3 py-1.5 rounded bg-indigo-600 text-white hover:bg-indigo-700 text-sm"
        >
          New poll
        </Link>
      </div>

      {loading && <p className="text-gray-500">Loading…</p>}
      {error && <p className="text-red-600 text-sm">{error}</p>}

      {!loading && polls.length === 0 && (
        <p className="text-gray-500">You haven't created any polls yet.</p>
      )}

      <ul className="space-y-3">
        {polls.map((poll) => (
          <li
            key={poll.id}
            className="bg-white border border-gray-200 rounded p-4 flex items-center justify-between"
          >
            <div>
              <p className="font-medium text-gray-900">{poll.title}</p>
              <p className="text-xs text-gray-500">{poll.options.length} options</p>
            </div>
            <div className="flex gap-2">
              <button
                onClick={() => copyLink(poll.id)}
                className="text-sm px-3 py-1.5 rounded bg-gray-100 hover:bg-gray-200 text-gray-700"
              >
                Copy link
              </button>
              <Link
                to={`/poll/${poll.id}`}
                className="text-sm px-3 py-1.5 rounded bg-indigo-50 hover:bg-indigo-100 text-indigo-700"
              >
                View
              </Link>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
