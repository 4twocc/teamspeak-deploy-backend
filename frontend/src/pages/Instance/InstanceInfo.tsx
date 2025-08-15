// src/pages/Instance/InstanceInfo.tsx
import React from 'react';
import { Card } from '@/components/ui/card';

interface InstanceInfoProps {
  instanceId: string;
}

const InstanceInfo: React.FC<InstanceInfoProps> = ({ instanceId }) => {
  // 这里应该从API获取实际的实例信息数据
  const instanceInfo = {
    id: instanceId,
    name: 'TeamSpeak Server',
    status: 'running',
    version: '3.13.7',
    ip: '192.168.1.100',
    port: 9987,
    created_at: '2023-01-15 10:30:00',
    updated_at: '2023-01-15 10:30:00',
  };

  const statusMap: Record<string, { text: string; color: string }> = {
    running: { text: '运行中', color: 'text-green-600' },
    stopped: { text: '已停止', color: 'text-red-600' },
    pending: { text: '启动中', color: 'text-yellow-600' },
  };

  return (
    <Card className="p-6">
      <h3 className="text-lg font-semibold mb-4">实例信息</h3>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label className="text-sm font-medium text-gray-500">实例ID</label>
          <p className="mt-1">{instanceInfo.id}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">实例名称</label>
          <p className="mt-1">{instanceInfo.name}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">状态</label>
          <p className={`mt-1 ${statusMap[instanceInfo.status]?.color || 'text-gray-600'}`}>
            {statusMap[instanceInfo.status]?.text || instanceInfo.status}
          </p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">版本</label>
          <p className="mt-1">{instanceInfo.version}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">IP地址</label>
          <p className="mt-1">{instanceInfo.ip}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">端口</label>
          <p className="mt-1">{instanceInfo.port}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">创建时间</label>
          <p className="mt-1">{instanceInfo.created_at}</p>
        </div>
        <div>
          <label className="text-sm font-medium text-gray-500">更新时间</label>
          <p className="mt-1">{instanceInfo.updated_at}</p>
        </div>
      </div>
    </Card>
  );
};

export default InstanceInfo;
