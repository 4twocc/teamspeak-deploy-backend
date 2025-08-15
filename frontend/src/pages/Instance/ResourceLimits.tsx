// src/pages/Instance/ResourceLimits.tsx
import React from 'react';
import { Card } from '@/components/ui/card';

const ResourceLimits: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  // 这里应该从API获取实际的资源限制数据
  const limits = {
    cpu_limit: 100, // CPU限制（百分比）
    memory_limit: 512, // 内存限制（MB）
    disk_limit: 5120, // 磁盘限制（MB）
  };

  return (
    <Card className="p-6">
      <h3 className="text-lg font-semibold mb-4">资源限制</h3>
      <div className="space-y-4">
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">CPU 限制</span>
            <span className="text-sm text-gray-500">{limits.cpu_limit}%</span>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">内存限制</span>
            <span className="text-sm text-gray-500">{limits.memory_limit} MB</span>
          </div>
        </div>
        
        <div>
          <div className="flex justify-between mb-1">
            <span className="text-sm font-medium">磁盘限制</span>
            <span className="text-sm text-gray-500">{limits.disk_limit} MB</span>
          </div>
        </div>
      </div>
    </Card>
  );
};

export default ResourceLimits;
