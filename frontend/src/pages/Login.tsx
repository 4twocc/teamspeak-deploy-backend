import React, { useState } from 'react'

const Login: React.FC = () => {
	const [username, setUsername] = useState('')
	const [password, setPassword] = useState('')
	const [error, setError] = useState('')

	const handleLogin = async (e: React.FormEvent) => {
		e.preventDefault()
		setError('')
		const res = await fetch('/api/auth/login', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username, password })
		})
		if (res.ok) {
			const data = await res.json()
			localStorage.setItem('token', data.token)
			window.location.href = '/dashboard'
		} else {
			setError('用户名或密码错误')
		}
	}

	return (
		<div className="flex flex-col items-center justify-center min-h-screen">
			<form onSubmit={handleLogin} className="bg-white p-6 rounded shadow-md w-80">
				<h2 className="text-2xl mb-4">登录</h2>
				<input className="border p-2 mb-2 w-full" placeholder="用户名" value={username} onChange={e => setUsername(e.target.value)} />
				<input className="border p-2 mb-2 w-full" type="password" placeholder="密码" value={password} onChange={e => setPassword(e.target.value)} />
				{error && <div className="text-red-500 mb-2">{error}</div>}
				<button className="bg-blue-500 text-black w-full py-2 rounded" type="submit">
					登录
				</button>
			</form>
		</div>
	)
}

export default Login
