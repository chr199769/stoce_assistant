import React, { useState } from 'react';
import { Card, Form, Input, Select, Button, Table, message, InputNumber, Space, Row, Col } from 'antd';
import { getGraphEntityProfile, getGraphNeighborhood, type GraphEntity, type GraphRelation } from '../api';

const entityTypeOptions = [
  { label: 'STOCK', value: 'STOCK' },
  { label: 'COMPANY', value: 'COMPANY' },
  { label: 'INDUSTRY', value: 'INDUSTRY' },
  { label: 'SECTOR', value: 'SECTOR' },
  { label: 'EVENT', value: 'EVENT' },
  { label: 'POLICY', value: 'POLICY' },
  { label: 'MACRO', value: 'MACRO' },
  { label: 'FLOW', value: 'FLOW' },
];

const relationTypeOptions = [
  { label: 'BELONGS_TO', value: 'BELONGS_TO' },
  { label: 'AFFECTS', value: 'AFFECTS' },
  { label: 'BENEFITS', value: 'BENEFITS' },
  { label: 'FLOWS_INTO', value: 'FLOWS_INTO' },
  { label: 'SUPPLIES', value: 'SUPPLIES' },
];

const GraphEntities: React.FC = () => {
  const [entity, setEntity] = useState<GraphEntity | undefined>();
  const [entities, setEntities] = useState<GraphEntity[]>([]);
  const [relations, setRelations] = useState<GraphRelation[]>([]);
  const [loading, setLoading] = useState(false);

  const handleSearch = async (values: any) => {
    setLoading(true);
    try {
      const relationTypes = (values.relation_types || []).join(',');
      const [profileRes, neighborhoodRes] = await Promise.all([
        getGraphEntityProfile({
          entity_type: values.entity_type,
          entity_id: values.entity_id,
          start_time: values.start_time,
          end_time: values.end_time,
        }),
        getGraphNeighborhood({
          entity_type: values.entity_type,
          entity_id: values.entity_id,
          relation_types: relationTypes || undefined,
          depth: values.depth,
          start_time: values.start_time,
          end_time: values.end_time,
          max_edges: values.max_edges,
        }),
      ]);
      setEntity(profileRes.entity);
      setEntities(neighborhoodRes.entities || []);
      setRelations(neighborhoodRes.relations || []);
    } catch (e) {
      console.error(e);
      message.error('查询失败');
    } finally {
      setLoading(false);
    }
  };

  const entityColumns = [
    { title: 'ID', dataIndex: 'id', key: 'id' },
    { title: '类型', dataIndex: 'type', key: 'type' },
    { title: '名称', dataIndex: 'name', key: 'name' },
    {
      title: '属性',
      dataIndex: 'attributes',
      key: 'attributes',
      render: (val: Record<string, string>) => (val ? JSON.stringify(val) : '-'),
    },
  ];

  const relationColumns = [
    {
      title: '来源',
      key: 'source',
      render: (_: any, record: GraphRelation) => `${record.source.name || record.source.id} (${record.source.type})`,
    },
    {
      title: '目标',
      key: 'target',
      render: (_: any, record: GraphRelation) => `${record.target.name || record.target.id} (${record.target.type})`,
    },
    { title: '关系类型', dataIndex: 'type', key: 'type' },
    { title: '强度', dataIndex: 'strength', key: 'strength' },
    {
      title: '证据摘要',
      key: 'evidence',
      render: (_: any, record: GraphRelation) => record.evidence?.summary || '-',
    },
    {
      title: '证据可信度',
      key: 'confidence',
      render: (_: any, record: GraphRelation) => record.evidence?.confidence ?? '-',
    },
  ];

  return (
    <Card title="图谱实体与关系检索">
      <Card size="small" style={{ marginBottom: 16 }}>
        <Form
          layout="vertical"
          initialValues={{ entity_type: 'STOCK', depth: 1, max_edges: 200 }}
          onFinish={handleSearch}
        >
          <Row gutter={[16, 4]} align="bottom" wrap>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item name="entity_type" label="实体类型" rules={[{ required: true, message: '请选择实体类型' }]}>
                <Select options={entityTypeOptions} allowClear showSearch style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={8}>
              <Form.Item name="entity_id" label="实体ID" rules={[{ required: true, message: '请输入实体ID' }]}>
                <Input placeholder="例如 sh601611 / sz000001" maxLength={15} />
              </Form.Item>
            </Col>
            <Col xs={24} md={10} lg={8}>
              <Form.Item name="relation_types" label="关系类型">
                <Select
                  mode="multiple"
                  options={relationTypeOptions}
                  allowClear
                  maxTagCount="responsive"
                  placeholder="不选则查询全部关系"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6} lg={4}>
              <Form.Item name="depth" label="深度">
                <InputNumber min={1} max={5} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6} lg={4}>
              <Form.Item name="max_edges" label="最大边数">
                <InputNumber min={1} max={5000} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={5}>
              <Form.Item name="start_time" label="开始时间">
                <InputNumber min={0} placeholder="时间戳(秒)" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={5}>
              <Form.Item name="end_time" label="结束时间">
                <InputNumber min={0} placeholder="时间戳(秒)" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={4} lg={2}>
              <Form.Item label=" " colon={false}>
                <Button type="primary" htmlType="submit" loading={loading} style={{ width: '100%' }}>
                  查询
                </Button>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>

      <Space direction="vertical" style={{ width: '100%' }} size="large">
        <Card size="small" title="实体信息">
          <Table
            dataSource={entity ? [entity] : []}
            columns={entityColumns}
            rowKey="id"
            pagination={false}
            loading={loading}
          />
        </Card>
        <Card size="small" title="邻居实体">
          <Table
            dataSource={entities}
            columns={entityColumns}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: true }}
          />
        </Card>
        <Card size="small" title="关系列表">
          <Table
            dataSource={relations}
            columns={relationColumns}
            rowKey={(record, index) => `${record.source.id}-${record.target.id}-${record.type}-${index}`}
            loading={loading}
            pagination={{ pageSize: 10, showSizeChanger: true }}
          />
        </Card>
      </Space>
    </Card>
  );
};

export default GraphEntities;
