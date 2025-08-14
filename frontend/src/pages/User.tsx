import React, { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectTrigger, SelectContent, SelectItem, SelectValue } from '@/components/ui/select'

type User = { username: string; role: string }

const Users: React.FC = () => {
	const [users, setUsers] = useState<User[]>([])
	const [username, setUsername] = useState('')
	const [role, setRole] = useState('user')

	const fetchUsers = () => {
		fetch('/api/users/list')
			.then(res => res.json())
			.then(setUsers)
	}

	useEffect(() => {
		fetchUsers()
	}, [])

	const addUser = async () => {
		if (!username) {
			toast('用户名不能为空')
			return
		}
		await fetch('/api/users/add', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ username, role })
		})
		setUsername('')
		setRole('user')
		fetchUsers()
		toast('添加成功', {
			description: `用户 ${username} 已添加`
		})
	}

	const removeUser = async (username: string) => {
		await fetch(`/api/users/remove?username=${username}`, { method: 'POST' })
		fetchUsers()
		toast('删除成功', {
			description: `用户 ${username} 已删除`
		})
	}

	return (
		<div className="p-8 flex flex-col items-center">
			<Card className="w-full max-w-xl p-6 space-y-4">
				<h1 className="text-xl font-bold">用户管理</h1>
				<div className="flex gap-2">
					<Input placeholder="用户名" value={username} onChange={e => setUsername(e.target.value)} />
					<Select value={role} onValueChange={setRole}>
						<SelectTrigger className="w-32">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="user">普通用户</SelectItem>
							<SelectItem value="admin">管理员</SelectItem>
						</SelectContent>
					</Select>
					<Button onClick={addUser}>添加</Button>
				</div>
				<ul>
					{users.map(u => (
						<li key={u.username} className="flex justify-between items-center border-b py-2">
							<span>
								{u.username}（{u.role}）
							</span>
							<Button variant="destructive" onClick={() => removeUser(u.username)}>
								删除
							</Button>
						</li>
					))}
				</ul>
			</Card>
		</div>
	)
}

export default Users
