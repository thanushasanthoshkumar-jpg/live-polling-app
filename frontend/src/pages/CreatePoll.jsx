import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import api, { getErrorMessage } from '../api.js'

export default function CreatePoll() {
  const [title, setTitle] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  function updateOption(index, value) {
    setOptions((prev) => prev.map((o, i) => (i === index ? value : o)))
  }

  function addOption() {
    if (options.length >= 10) return
    setOptions((prev) => [...prev, ''])
  }

  function removeOption(index) {
    if (options.length <= 2) return
    setOptions((prev) => prev.filter((_, i) => i !== index))
  }

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')

    const cleaned = options.map((o) => o.trim()).filter(Boolean)
    if (cleaned.length < 2) {
      setError('Provide at least 2 non-empty options.')
      return
    }
    const unique = new Set(cleaned.map((o) => o.toLowerCase()))
    if (unique.size !== cleaned.length) {
      setError('Options must be unique.')
      return
    }

    setLoading(true)
    try {
      const res = await api.post('/polls', { title: title.trim(), options: cleaned })
      navigate(`/poll/${res.data.id}`)
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-lg mx-auto px-4 py-10">
      <h2 className="text-2xl font-bold mb-6 text-gray-900">Create a poll</h2>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm text-gray-700 mb-1">Question</label>
          <input
            type="text"
            required
            minLength={3}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="e.g. What should we build next?"
            className="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>

        <div>
          <label className="block text-sm text-gray-700 mb-1">Options</label>
          <div className="space-y-2">
            {options.map((opt, i) => (
              <div key={i} className="flex gap-2">
                <input
                  type="text"
                  required
                  value={opt}
                  onChange={(e) => updateOption(i, e.target.value)}
                  placeholder={`Option ${i + 1}`}
                  className="flex-1 border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                />
                {options.length > 2 && (
                  <button
                    type="button"
                    onClick={() => removeOption(i)}
                    className="px-3 rounded bg-gray-100 hover:bg-gray-200 text-gray-600"
                  >
                    ✕
                  </button>
                )}
              </div>
            ))}
          </div>
          {options.length < 10 && (
            <button
              type="button"
              onClick={addOption}
              className="mt-2 text-sm text-indigo-600 hover:underline"
            >
              + Add option
            </button>
          )}
        </div>

        {error && <p className="text-sm text-red-600">{error}</p>}

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-indigo-600 text-white py-2 rounded hover:bg-indigo-700 disabled:opacity-50"
        >
          {loading ? 'Creating…' : 'Create poll'}
        </button>
      </form>
    </div>
  )
}
