// frontend/src/pages/Dashboard.tsx
import React, { useState, useEffect } from 'react'


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
      message.error('获取部署状态失败: ' + error)
    }
  }

  const startDeploy = async () => {
    try {
      setLoading(true)
      const response = await fetch('/api/v1/deploy/start', {
        method: 'POST'
      })

      if (response.ok) {
        message.success('部署已启动')
        fetchDeployStatus()
      } else {
        const error = await response.json()
        message.error('启动部署失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('启动部署失败: ' + error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'success'
      case 'running': return 'processing'
      case 'error': return 'error'
      default: return 'default'
    }
  }

  return (
    <div className="p-8">
      <Title level={2}>控制面板</Title>
      
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Card title="一键部署 TeamSpeak">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Space>
              <Text strong>部署状态:</Text>
              <Tag color={getStatusColor(deployStatus.status)}>
                {deployStatus.status}
              </Tag>
            </Space>
            
            <Text>{deployStatus.message}</Text>
            
            <Space>
              <Button 
                type="primary" 
                icon={<PlayCircleOutlined />} 
                onClick={startDeploy}
                loading={loading}
                disabled={deployStatus.status === 'running'}
              >
                开始部署
              </Button>
              <Button 
                icon={<ReloadOutlined />} 
                onClick={fetchDeployStatus}
              >
                刷新状态
              </Button>
            </Space>
          </Space>
        </Card>
        
        <Card title="系统监控">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text>系统监控功能正在开发中...</Text>
            <Progress percent={70} status="active" />
          </Space>
        </Card>
        
        <Card title="使用说明">
          <Space direction="vertical">
            <Text>
              1. 点击"开始部署"按钮一键部署 TeamSpeak 服务器
            </Text>
            <Text>
              2. 部署完成后可以在"实例管理"页面管理您的 TeamSpeak 实例
            </Text>
            <Text>
              3. 可以通过"系统监控"查看服务器运行状态
            </Text>
          </Space>
        </Card>
      </Space>
    </div>
  )
}

export default Dashboard
