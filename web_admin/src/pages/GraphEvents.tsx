import React, { useState } from 'react';
import { Card, Form, Input, Select, Button, Table, message, InputNumber, Row, Col } from 'antd';
import { searchGraphEvents, type GraphEvent } from '../api';

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

const eventTypeOptions = [
  { label: 'NEWS', value: 'NEWS' },
  { label: 'POLICY', value: 'POLICY' },
  { label: 'REPORT', value: 'REPORT' },
  { label: 'RUMOR', value: 'RUMOR' },
  { label: 'CAPITAL', value: 'CAPITAL' },
];

const GraphEvents: React.FC = () => {
  const [events, setEvents] = useState<GraphEvent[]>([]);
  const [loading, setLoading] = useState(false);

  const handleSearch = async (values: any) => {
    setLoading(true);
    try {
      const eventTypes = (values.event_types || []).join(',');
      const res = await searchGraphEvents({
        entity_type: values.entity_type,
        entity_id: values.entity_id,
        event_types: eventTypes || undefined,
        start_time: values.start_time,
        end_time: values.end_time,
        min_confidence: values.min_confidence,
        limit: values.limit,
      });
      setEvents(res.events || []);
    } catch (e) {
      console.error(e);
      message.error('查询失败');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    { title: '事件类型', dataIndex: 'type', key: 'type' },
    { title: '来源', dataIndex: 'source', key: 'source' },
    { title: '可信度', dataIndex: 'confidence', key: 'confidence' },
    { title: '影响方向', dataIndex: 'impact_direction', key: 'impact_direction' },
    { title: '影响强度', dataIndex: 'impact_strength', key: 'impact_strength' },
    { title: '时间戳', dataIndex: 'timestamp', key: 'timestamp' },
    {
      title: '关联实体',
      key: 'entities',
      render: (_: any, record: GraphEvent) => record.entities.map((e) => `${e.name || e.id}(${e.type})`).join('，'),
    },
  ];

  return (
    <Card title="事件查询与可信度筛选">
      <Card size="small" style={{ marginBottom: 16 }}>
        <Form
          layout="vertical"
          initialValues={{ entity_type: 'STOCK', limit: 20, min_confidence: 0 }}
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
              <Form.Item name="event_types" label="事件类型">
                <Select
                  mode="multiple"
                  options={eventTypeOptions}
                  allowClear
                  maxTagCount="responsive"
                  placeholder="不选则查询全部事件"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6} lg={4}>
              <Form.Item name="min_confidence" label="最低可信度">
                <InputNumber min={0} max={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6} lg={4}>
              <Form.Item name="limit" label="数量上限">
                <InputNumber min={1} max={200} style={{ width: '100%' }} />
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

      <Table
        dataSource={events}
        columns={columns}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 10, showSizeChanger: true }}
      />
    </Card>
  );
};

export default GraphEvents;
