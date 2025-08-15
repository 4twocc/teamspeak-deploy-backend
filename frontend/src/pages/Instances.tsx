import React, { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Trash2, Play, Square, RotateCw, Plus } from 'lucide-react'
import type { TSInstance} from '@/types/Instance'

const Instances: React.FC = () => {
  const [instances, setInstances] = useState<TSInstance[]>([])
  const [_loading, setLoading] = useState(false)
  const [name, setName] = useState('')
  const [serverName, setServerName] = useState('')
  const [maxClients, setMaxClients] = useState(32)

  useEffect(() => {
    fetchInstances()
  }, [])

  const fetchInstances = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/v1/instances')
      const data = await response.json()
      setInstances(data.data || [])
    } catch (error) {
      toast.error('获取实例列表失败: ' + error)
    } finally {
      setLoading(false)
    }
  }

  const createInstance = async () => {
    if (!name || !serverName) {
      toast.warning('请填写实例名和服务器名')
      return
    }

    try {
      const response = await fetch('/api/v1/instances', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name,
          server_name: serverName,
          max_clients: maxClients,
          version: '3.13.7',
          host: 'localhost',
          voice_port: 9987,
          file_port: 30033,
          query_port: 10011,
          server_port: 2010,
          query_admin_password: 'admin123',
          log_queries: true,
          log_client_cmds: true
        })
      })

      if (response.ok) {
        toast.success('实例创建成功')
        setName('')
        setServerName('')
        setMaxClients(32)
        fetchInstances()
      } else {
        const error = await response.json()
        toast.error('创建实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('创建实例失败: ' + error)
    }
  }

  const deleteInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}`, {
        method: 'DELETE'
      })

      if (response.ok) {
        toast.success('实例删除成功')
        fetchInstances()
      } else {
        const error = await response.json()
        toast.error('删除实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('删除实例失败: ' + error)
    }
  }

  const startInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/start`, {
        method: 'POST'
      })

      if (response.ok) {
        toast.success('启动实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        toast.error('启动实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('启动实例失败: ' + error)
    }
  }

  const stopInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/stop`, {
        method: 'POST'
      })

      if (response.ok) {
        toast.success('停止实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        toast.error('停止实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('停止实例失败: ' + error)
    }
  }

  const restartInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/restart`, {
        method: 'POST'
      })

      if (response.ok) {
        toast.success('重启实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        toast.error('重启实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('重启实例失败: ' + error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-500'
      case 'stopped': return 'bg-gray-500'
      case 'starting':
      case 'stopping': return 'bg-blue-500'
      case 'error': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  return (
    <div className="p-8">
      <h2 className="text-2xl font-bold mb-6">TeamSpeak 实例管理</h2>
      
      <Card className="p-6 mb-6">
        <h3 className="text-xl font-semibold mb-4">创建新实例</h3>
        <div className="flex flex-wrap gap-2">
          <Input 
            placeholder="实例名" 
            value={name} 
            onChange={e => setName(e.target.value)} 
            className="w-48"
          />
          <Input 
            placeholder="服务器名" 
            value={serverName} 
            onChange={e => setServerName(e.target.value)} 
            className="w-48"
          />
          <Input 
            type="number"
            placeholder="最大客户端数" 
            value={maxClients} 
            onChange={e => setMaxClients(Number(e.target.value))} 
            className="w-32"
          />
          <Button 
            onClick={createInstance}
            className="flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            创建实例
          </Button>
        </div>
      </Card>

      <Card className="p-6">
        <h3 className="text-xl font-semibold mb-4">实例列表</h3>
        <div className="space-y-4">
          {instances.map(instance => (
            <div key={instance.id} className="border rounded-lg p-4 flex flex-col sm:flex-row justify-between gap-4">
              <div>
                <div className="flex items-center gap-2 mb-2">
                  <h4 className="font-medium">{instance.name}</h4>
                  <span className={`px-2 py-1 rounded text-white text-xs ${getStatusColor(instance.status)}`}>
                    {instance.status}
                  </span>
                </div>
                <div className="text-sm text-gray-600 space-y-1">
                  <p>
                    服务器: {instance.config.server_name} | 
                    语音端口: {instance.config.voice_port} | 
                    版本: {instance.version}
                  </p>
                  <p>
                    创建时间: {new Date(instance.created_at).toLocaleString()}
                  </p>
                </div>
              </div>
              <div className="flex flex-wrap gap-2">
                <Button 
                  size="sm"
                  onClick={() => startInstance(instance.id)}
                  disabled={instance.status !== 'stopped'}
                  className="flex items-center gap-1"
                >
                  <Play className="w-4 h-4" />
                  启动
                </Button>
                <Button 
                  variant="outline"
                  size="sm"
                  onClick={() => stopInstance(instance.id)}
                  disabled={instance.status !== 'running'}
                  className="flex items-center gap-1"
                >
                  <Square className="w-4 h-4" />
                  停止
                </Button>
                <Button 
                  variant="outline"
                  size="sm"
                  onClick={() => restartInstance(instance.id)}
                  disabled={instance.status !== 'running' && instance.status !== 'error'}
                  className="flex items-center gap-1"
                >
                  <RotateCw className="w-4 h-4" />
                  重启
                </Button>
                <Button 
                  variant="destructive"
                  size="sm"
                  onClick={() => deleteInstance(instance.id)}
                  className="flex items-center gap-1"
                >
                  <Trash2 className="w-4 h-4" />
                  删除
                </Button>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )
}

export default Instances
