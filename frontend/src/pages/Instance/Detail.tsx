// src/pages/Instance/Detail.tsx
import React, { useState } from 'react';
import InstanceInfo from './InstanceInfo';
import InstanceLogs from './InstanceLogs';
import InstanceResources from './InstanceResources';
import ResourceLimits from './ResourceLimits';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { 
  Play, 
  Square, 
  RotateCw, 
  Trash2,
  AlertTriangle
} from 'lucide-react';
import { toast } from 'sonner';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { 
  startInstance, 
  stopInstance, 
  restartInstance, 
  deleteInstance 
} from '@/api/instance';

const InstanceDetail: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  const [activeTab, setActiveTab] = useState('info');
  const [loading, setLoading] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const handleStart = async () => {
    try {
      setLoading(true);
      await startInstance(instanceId);
      toast.success('实例启动成功');
      // 刷新当前标签页数据
      window.dispatchEvent(new Event('refresh-instance-data'));
    } catch (error) {
      console.error(error);
      toast.error('启动实例失败');
    } finally {
      setLoading(false);
    }
  };

  const handleStop = async () => {
    try {
      setLoading(true);
      await stopInstance(instanceId);
      toast.success('实例停止成功');
      window.dispatchEvent(new Event('refresh-instance-data'));
    } catch (error) {
      console.error(error);
      toast.error('停止实例失败');
    } finally {
      setLoading(false);
    }
  };

  const handleRestart = async () => {
    try {
      setLoading(true);
      await restartInstance(instanceId);
      toast.success('实例重启成功');
      window.dispatchEvent(new Event('refresh-instance-data'));
    } catch (error) {
      console.error(error);
      toast.error('重启实例失败');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    try {
      setLoading(true);
      await deleteInstance(instanceId);
      toast.success('实例删除成功');
      // 返回实例列表页面
      window.location.hash = '#/instances';
    } catch (error) {
      console.error(error);
      toast.error('删除实例失败');
    } finally {
      setLoading(false);
      setShowDeleteDialog(false);
    }
  };

  return (
    <Card className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-2xl font-bold">实例详情</h2>
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            size="sm"
            onClick={handleStart}
            disabled={loading}
            className="flex items-center gap-2"
          >
            <Play className="w-4 h-4" />
            启动
          </Button>
          <Button 
            variant="outline" 
            size="sm"
            onClick={handleStop}
            disabled={loading}
            className="flex items-center gap-2"
          >
            <Square className="w-4 h-4" />
            停止
          </Button>
          <Button 
            variant="outline" 
            size="sm"
            onClick={handleRestart}
            disabled={loading}
            className="flex items-center gap-2"
          >
            <RotateCw className="w-4 h-4" />
            重启
          </Button>
          <Button 
            variant="destructive" 
            size="sm"
            onClick={() => setShowDeleteDialog(true)}
            disabled={loading}
            className="flex items-center gap-2"
          >
            <Trash2 className="w-4 h-4" />
            删除
          </Button>
        </div>
      </div>

      <div className="border-b border-gray-200 mb-4">
        <nav className="-mb-px flex space-x-8">
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'info'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('info')}
          >
            实例信息
          </button>
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'resources'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('resources')}
          >
            资源监控
          </button>
          <button
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'logs'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setActiveTab('logs')}
          >
            日志
          </button>
        </nav>
      </div>
      
      <div className="py-4">
        {activeTab === 'info' && <InstanceInfo instanceId={instanceId} />}
        {activeTab === 'resources' && (
          <div>
            <InstanceResources instanceId={instanceId} />
            <div className="mt-4">
              <ResourceLimits instanceId={instanceId} />
            </div>
          </div>
        )}
        {activeTab === 'logs' && <InstanceLogs instanceId={instanceId} />}
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <AlertTriangle className="text-yellow-500" />
              确认删除实例
            </AlertDialogTitle>
            <AlertDialogDescription>
              此操作将永久删除实例 "{instanceId}"。删除后将无法恢复实例数据。
              <br />
              <br />
              请确认是否继续删除？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleDelete}
              className="bg-red-600 hover:bg-red-700"
            >
              确认删除
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Card>
  );
};

export default InstanceDetail;