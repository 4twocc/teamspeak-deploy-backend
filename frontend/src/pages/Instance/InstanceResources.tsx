// src/pages/Instance/InstanceResources.tsx
import React, { useEffect, useState } from 'react';
import { getInstanceResources } from '@/services/instance';
import type { ResourceUsage } from '@/types/Instance';

const InstanceResources: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [loading, setLoading] = useState(false);
  const [resources, setResources] = useState<ResourceUsage | null>(null);

  const fetchResources = async () => {
    try {
      setLoading(true);
      const data = await getInstanceResources(instanceId);
      setResources(data);
    } catch (error) {
      message.error('获取资源使用情况失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchResources();
    const timer = setInterval(fetchResources, 30000); // 每30秒刷新一次
    
    return () => clearInterval(timer);
  }, [instanceId]);

  const columns = [
    {
      title: '资源类型',
      dataIndex: 'resource',
      key: 'resource',
    },
    {
      title: '使用量',
      dataIndex: 'usage',
      key: 'usage',
      render: (text: string, record: any) => {
        if (record.type === 'cpu') {
          return <Progress percent={resources?.cpu_percent} status={resources?.cpu_percent > 80 ? 'exception' : 'normal'} />;
        } else if (record.type === 'memory') {
          return `${(resources?.memory_mb || 0).toFixed(2)} MB (${resources?.memory_percent.toFixed(2)}%)`;
        } else if (record.type === 'disk') {
          return `${resources?.disk_usage_mb} MB`;
        }
        return text;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (text: string, record: any) => {
        let isOverLimit = false;
        if (record.type === 'cpu' && resources?.cpu_percent > 80) {
          isOverLimit = true;
        } else if (record.type === 'memory' && resources?.memory_percent > 80) {
          isOverLimit = true;
        }
        
        return isOverLimit ? <Tag color="error">超限</Tag> : <Tag color="success">正常</Tag>;
      },
    },
  ];

  const dataSource = [
    { key: '1', resource: 'CPU', type: 'cpu' },
    { key: '2', resource: '内存', type: 'memory' },
    { key: '3', resource: '磁盘', type: 'disk' },
  ];

  return (
    <Card 
      title="资源使用情况" 
      extra={
        <Button 
          icon={<ReloadOutlined />} 
          onClick={fetchResources} 
          loading={loading}
        >
          刷新
        </Button>
      }
    >
      <Table 
        columns={columns} 
        dataSource={dataSource} 
        pagination={false}
        loading={loading}
      />
    </Card>
  );
};

export default InstanceResources;
