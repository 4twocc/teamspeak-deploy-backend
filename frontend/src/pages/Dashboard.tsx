// frontend/src/pages/Dashboard.tsx
import React, { useState, useEffect } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'

const Dashboard: React.FC = () => {
  const [deployStatus, setDeployStatus] = useState({
    status: 'idle',
    message: 'Ready to deploy'
  })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    fetchDeployStatus()
  }, [])

  const fetchDeployStatus = async () => {
    try {
      const response = await fetch('/api/v1/deploy/status')
      const data = await response.json()
      setDeployStatus(data.data || { status: 'idle', message: 'Unknown status' })
    } catch (error) {
      toast.error('获取部署状态失败: ' + error)
    }
  }

  const startDeploy = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/v1/deploy/start', {
        method: 'POST'
      })

      if (response.ok) {
        toast.success('部署已启动')
        fetchDeployStatus()
      } else {
        const error = await response.json()
        toast.error('启动部署失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      toast.error('启动部署失败: ' + error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-500'
      case 'running': return 'bg-blue-500'
      case 'error': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  return (
    <div className="p-8">
      <h2 className="text-2xl font-bold mb-6">控制面板</h2>
      
      <div className="space-y-6">
        <Card className="p-6">
          <h3 className="text-xl font-semibold mb-4">一键部署 TeamSpeak</h3>
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <span className="font-medium">部署状态:</span>
              <span className={`px-2 py-1 rounded text-white text-sm ${getStatusColor(deployStatus.status)}`}>
                {deployStatus.status}
              </span>
            </div>
            
            <p>{deployStatus.message}</p>
            
            <div className="flex gap-2">
              <Button 
                onClick={startDeploy}
                disabled={loading || deployStatus.status === 'running'}
              >
                {loading ? '部署中...' : '开始部署'}
              </Button>
              <Button 
                variant="outline"
                onClick={fetchDeployStatus}
              >
                刷新状态
              </Button>
            </div>
          </div>
        </Card>
        
        <Card className="p-6">
          <h3 className="text-xl font-semibold mb-4">系统监控</h3>
          <div className="space-y-4">
            <p>系统监控功能正在开发中...</p>
            <div className="w-full bg-gray-200 rounded-full h-2.5">
              <div className="bg-blue-600 h-2.5 rounded-full" style={{ width: '70%' }}></div>
            </div>
          </div>
        </Card>
        
        <Card className="p-6">
          <h3 className="text-xl font-semibold mb-4">使用说明</h3>
          <ul className="list-disc pl-5 space-y-2">
            <li>点击"开始部署"按钮一键部署 TeamSpeak 服务器</li>
            <li>部署完成后可以在"实例管理"页面管理您的 TeamSpeak 实例</li>
            <li>可以通过"系统监控"查看服务器运行状态</li>
          </ul>
        </Card>
      </div>
    </div>
  )
}

export default Dashboard
