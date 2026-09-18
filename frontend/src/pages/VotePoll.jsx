import React, { useEffect, useRef, useState, useCallback } from 'react'
import { useParams } from 'react-router-dom'
import api, { getErrorMessage, WS_BASE_URL } from '../api.js'
import PollResults from '../components/PollResults.jsx'

export default function VotePoll() {
  const { id } = useParams()
  const [poll, setPoll] = useState(null)
  const [counts, setCounts] = useState({})
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [voting, setVoting] = useState(false)
  const [voted, setVoted] = useState(false)
  const [wsConnected, setWsConnected] = useState(false)
  const wsRef = useRef(null)

  const votedKey = `voted_poll_${id}`

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const res = await api.get(`/polls/${id}`)
        if (cancelled) return
        setPoll(res.data.poll)
        setCounts(res.data.counts || {})
      } catch (err) {
        if (!cancelled) setError(getErrorMessage(err))
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    if (localStorage.getItem(votedKey)) setVoted(true)
    return () => {
      cancelled = true
    }
  }, [id])

  // Connect to the WebSocket endpoint for real-time vote updates.
  useEffect(() => {
    if (!id) return

    let ws
    let retryTimer

    function connect() {
      ws = new WebSocket(`${WS_BASE_URL}/ws/polls/${id}`)
      wsRef.current = ws

      ws.onopen = () => setWsConnected(true)
      ws.onclose = () => {
        setWsConnected(false)
        // Auto-reconnect after a short delay if the tab is still open.
        retryTimer = setTimeout(connect, 2000)
      }
      ws.onerror = () => ws.close()
      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data)
          if (payload.poll_id === id && payload.counts) {
            setCounts(payload.counts)
          }
        } catch {
          // ignore malformed messages
        }
      }
    }

    connect()

    return () => {
      clearTimeout(retryTimer)
      wsRef.current?.close()
    }
  }, [id])

  const castVote = useCallback(
    async (optionId) => {
      if (voted || voting) return
      setVoting(true)
      setError('')
      try {
        const res = await api.post(`/polls/${id}/vote`, { option_id: optionId })
        setCounts(res.data.counts)
        setVoted(true)
        localStorage.setItem(votedKey, optionId)
      } catch (err) {
        setError(getErrorMessage(err))
      } finally {
        setVoting(false)
      }
    },
    [id, voted, voting, votedKey],
  )

  if (loading) {
    return <div className="max-w-lg mx-auto px-4 py-10 text-gray-500">Loading poll…</div>
  }

  if (error && !poll) {
    return <div className="max-w-lg mx-auto px-4 py-10 text-red-600">{error}</div>
  }

  return (
    <div className="max-w-lg mx-auto px-4 py-10">
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <div className="flex items-center justify-between mb-1">
          <h2 className="text-xl font-bold text-gray-900">{poll.title}</h2>
          <span
            className={`text-xs px-2 py-0.5 rounded-full ${
              wsConnected ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-100 text-gray-500'
            }`}
            title={wsConnected ? 'Live updates connected' : 'Reconnecting…'}
          >
            {wsConnected ? '● Live' : '○ Connecting'}
          </span>
        </div>

        {!poll.is_active && (
          <p className="text-sm text-amber-600 mb-4">This poll is closed to new votes.</p>
        )}

        {!voted && poll.is_active ? (
          <div className="space-y-2 mt-4">
            {poll.options.map((opt) => (
              <button
                key={opt.id}
                onClick={() => castVote(opt.id)}
                disabled={voting}
                className="w-full text-left border border-gray-300 rounded px-4 py-2.5 hover:border-indigo-500 hover:bg-indigo-50 disabled:opacity-50 transition-colors"
              >
                {opt.text}
              </button>
            ))}
          </div>
        ) : (
          <div className="mt-4">
            <PollResults options={poll.options} counts={counts} />
          </div>
        )}

        {error && <p className="text-sm text-red-600 mt-3">{error}</p>}

        {voted && (
          <p className="text-xs text-gray-400 mt-4">
            Thanks for voting — results update live as others vote.
          </p>
        )}
      </div>
    </div>
  )
}
