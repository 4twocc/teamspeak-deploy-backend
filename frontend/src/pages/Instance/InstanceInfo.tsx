// src/pages/Instance/InstanceInfo.tsx
import React, { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { RotateCw } from 'lucide-react';
import { toast } from 'sonner';
import { getInstance } from '@/api/instance';
import type { TSInstance } from '@/types/Instance';

interface InstanceInfoProps {
  instanceId: string;
}

const InstanceInfo: React.FC<InstanceInfoProps> = ({ instanceId }) => {
  const [loading, setLoading] = useState(false);
  const [instanceInfo, setInstanceInfo] = useState<TSInstance | null>(null);

  const fetchInstanceInfo = async () => {
    try {
      setLoading(true);
      const response = await getInstance(instanceId);
      setInstanceInfo(response.data);
    } catch (error) {
      console.error(error);
      toast.error('获取实例信息失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchInstanceInfo();
  }, [instanceId]);

  const statusMap: Record<string, { text: string; color: string }> = {
    running: { text: '运行中', color: 'text-green-600' },
    stopped: { text: '已停止', color: 'text-red-600' },
    starting: { text: '启动中', color: 'text-yellow-600' },
    stopping: { text: '停止中', color: 'text-orange-600' },
    error: { text: '错误', color: 'text-red-600' },
  };

  return (
    <Card className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold">实例信息</h3>
        <Button 
          variant="outline" 
          size="sm" 
          onClick={fetchInstanceInfo} 
          disabled={loading}
          className="flex items-center gap-2"
        >
          <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          刷新
        </Button>
      </div>
      
      {instanceInfo ? (
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
            <label className="text-sm font-medium text-gray-500">连接地址</label>
            <p className="mt-1">{instanceInfo.host}:{instanceInfo.config.voice_port}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">创建时间</label>
            <p className="mt-1">{new Date(instanceInfo.created_at).toLocaleString()}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">更新时间</label>
            <p className="mt-1">{new Date(instanceInfo.updated_at).toLocaleString()}</p>
          </div>
        </div>
      ) : (
        <div className="flex justify-center items-center h-32">
          <div className="text-gray-500">
            {loading ? '加载中...' : '暂无数据'}
          </div>
        </div>
      )}
    </Card>
  );
};

export default InstanceInfo;