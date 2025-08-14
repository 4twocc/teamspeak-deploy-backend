// src/pages/Instance/Detail.tsx
import React from 'react';
import InstanceInfo from './InstanceInfo';
import InstanceLogs from './InstanceLogs';
import InstanceResources from './InstanceResources';
import ResourceLimits from './ResourceLimits';

const { TabPane } = Tabs;

const InstanceDetail: React.FC<{ instanceId: string }> = ({ instanceId }) => {
  return (
    <Card>
      <Tabs defaultActiveKey="1">
        <TabPane tab="实例信息" key="1">
          <InstanceInfo instanceId={instanceId} />
        </TabPane>
        <TabPane tab="资源监控" key="2">
          <InstanceResources instanceId={instanceId} />
          <div style={{ marginTop: 16 }}>
            <ResourceLimits instanceId={instanceId} />
          </div>
        </TabPane>
        <TabPane tab="日志" key="3">
          <InstanceLogs instanceId={instanceId} />
        </TabPane>
      </Tabs>
    </Card>
  );
};

export default InstanceDetail;
