import React, { useEffect, useState } from 'react'
import type { TSInstance } from '@/types/Instance'

const Instances: React.FC = () => {
  const [instances, setInstances] = useState<TSInstance[]>([])
  const [loading, setLoading] = useState(false)
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
      message.error('获取实例列表失败: ' + error)
    } finally {
      setLoading(false)
    }
  }

  const createInstance = async () => {
    if (!name || !serverName) {
      message.warning('请填写实例名和服务器名')
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
        message.success('实例创建成功')
        setName('')
        setServerName('')
        setMaxClients(32)
        fetchInstances()
      } else {
        const error = await response.json()
        message.error('创建实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('创建实例失败: ' + error)
    }
  }

  const deleteInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}`, {
        method: 'DELETE'
      })

      if (response.ok) {
        message.success('实例删除成功')
        fetchInstances()
      } else {
        const error = await response.json()
        message.error('删除实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('删除实例失败: ' + error)
    }
  }

  const startInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/start`, {
        method: 'POST'
      })

      if (response.ok) {
        message.success('启动实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        message.error('启动实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('启动实例失败: ' + error)
    }
  }

  const stopInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/stop`, {
        method: 'POST'
      })

      if (response.ok) {
        message.success('停止实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        message.error('停止实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('停止实例失败: ' + error)
    }
  }

  const restartInstance = async (id: string) => {
    try {
      const response = await fetch(`/api/v1/instances/${id}/restart`, {
        method: 'POST'
      })

      if (response.ok) {
        message.success('重启实例请求已提交')
        fetchInstances()
      } else {
        const error = await response.json()
        message.error('重启实例失败: ' + (error.message || '未知错误'))
      }
    } catch (error) {
      message.error('重启实例失败: ' + error)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'success'
      case 'stopped': return 'default'
      case 'starting':
      case 'stopping': return 'processing'
      case 'error': return 'error'
      default: return 'default'
    }
  }

  return (
    <div className="p-8">
      <Title level={2}>TeamSpeak 实例管理</Title>
      
      <Card title="创建新实例" style={{ marginBottom: 24 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            <Input 
              placeholder="实例名" 
              value={name} 
              onChange={e => setName(e.target.value)} 
              style={{ width: 200 }}
            />
            <Input 
              placeholder="服务器名" 
              value={serverName} 
              onChange={e => setServerName(e.target.value)} 
              style={{ width: 200 }}
            />
            <Input 
              type="number"
              placeholder="最大客户端数" 
              value={maxClients} 
              onChange={e => setMaxClients(Number(e.target.value))} 
              style={{ width: 150 }}
            />
            <Button 
              type="primary" 
              icon={<PlusOutlined />} 
              onClick={createInstance}
            >
              创建实例
            </Button>
          </Space>
        </Space>
      </Card>

      <Card title="实例列表">
        <List
          loading={loading}
          dataSource={instances}
          renderItem={instance => (
            <List.Item
              actions={[
                <Button 
                  type="primary" 
                  icon={<PlayCircleOutlined />} 
                  size="small"
                  disabled={instance.status !== 'stopped'}
                  onClick={() => startInstance(instance.id)}
                >
                  启动
                </Button>,
                <Button 
                  icon={<StopOutlined />} 
                  size="small"
                  disabled={instance.status !== 'running'}
                  onClick={() => stopInstance(instance.id)}
                >
                  停止
                </Button>,
                <Button 
                  icon={<RedoOutlined />} 
                  size="small"
                  disabled={instance.status !== 'running' && instance.status !== 'error'}
                  onClick={() => restartInstance(instance.id)}
                >
                  重启
                </Button>,
                <Popconfirm
                  title="确认删除实例？"
                  description="此操作不可恢复"
                  onConfirm={() => deleteInstance(instance.id)}
                  okText="确认"
                  cancelText="取消"
                >
                  <Button icon={<DeleteOutlined />} size="small" danger>
                    删除
                  </Button>
                </Popconfirm>
              ]}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <span>{instance.name}</span>
                    <Tag color={getStatusColor(instance.status)}>
                      {instance.status}
                    </Tag>
                  </Space>
                }
                description={
                  <Space direction="vertical" size="small">
                    <Text type="secondary">
                      服务器: {instance.config.server_name} | 
                      语音端口: {instance.config.voice_port} | 
                      版本: {instance.version}
                    </Text>
                    <Text type="secondary">
                      创建时间: {new Date(instance.created_at).toLocaleString()}
                    </Text>
                  </Space>
                }
              />
            </List.Item>
          )}
        />
      </Card>
    </div>
  )
}

export default Instances
