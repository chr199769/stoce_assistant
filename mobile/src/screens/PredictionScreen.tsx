import React, { useState, useEffect, useCallback } from 'react';
import { View, StyleSheet, ScrollView, ActivityIndicator } from 'react-native';
import { Appbar, TextInput, Button, Card, Text, ProgressBar, HelperText, SegmentedButtons } from 'react-native-paper';
import { getPrediction } from '../api/stock';
import { PredictionResponse } from '../types';
import { useRoute, useNavigation } from '@react-navigation/native';
import FractalChart, { FractalData } from '../components/FractalChart';

const parseSummarySections = (summary: string) => {
  const evidenceMarker = '证据摘要:';
  const conflictMarker = '冲突说明:';
  let main = summary || '';
  let conflict = '';
  if (main.includes(evidenceMarker)) {
    const parts = main.split(evidenceMarker);
    main = parts[0].trim();
    const rest = parts.slice(1).join(evidenceMarker).trim();
    if (rest.includes(conflictMarker)) {
      const conflictParts = rest.split(conflictMarker);
      conflict = conflictParts.slice(1).join(conflictMarker).trim();
    }
  } else if (main.includes(conflictMarker)) {
    const parts = main.split(conflictMarker);
    main = parts[0].trim();
    conflict = parts.slice(1).join(conflictMarker).trim();
  }
  return { main, conflict };
};

const parseAnalysisSections = (analysis: string) => {
  const markers = ['智能体理由:', '智能体理由：', '智能体理由', '【智能体理由】'];
  let main = analysis || '';
  let reasons = '';
  for (const marker of markers) {
    if (main.includes(marker)) {
      const parts = main.split(marker);
      main = parts[0].trim();
      reasons = parts.slice(1).join(marker).trim();
      if (reasons.startsWith(':') || reasons.startsWith('：')) {
        reasons = reasons.slice(1).trim();
      }
      break;
    }
  }
  const reasonLines = reasons
    ? reasons
      .split(/\r?\n/)
      .map(line => line.replace(/^[-•\d.\s]+/, '').trim())
      .filter(Boolean)
    : [];
  return { main, reasonLines };
};

const PredictionScreen = () => {
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PredictionResponse | null>(null);
  const [error, setError] = useState('');
  const [model, setModel] = useState('glm-4-flash-250414');
  const [fractalData, setFractalData] = useState<FractalData | null>(null);
  const [stockName, setStockName] = useState<string>('');

  const route = useRoute();
  const navigation = useNavigation();

  const parseMetadata = (analysis: string) => {
    const parts = analysis.split('---METADATA---');
    if (parts.length > 1) {
      try {
        const metadataJson = parts[1].trim();
        const metadata = JSON.parse(metadataJson);
        if (metadata.fractal_data) {
          setFractalData(metadata.fractal_data);
        } else {
          setFractalData(null);
        }
        return parts[0].trim(); // Return cleaned analysis text
      } catch (e) {
        console.error("Failed to parse metadata", e);
      }
    }
    setFractalData(null);
    return analysis;
  };

  const handlePredict = useCallback(async (searchCode: string, currentModel: string) => {
    if (!searchCode) return;
    setLoading(true);
    setError('');
    setResult(null);
    setFractalData(null);

    try {
      const data = await getPrediction({
        code: searchCode,
        days: 3,
        include_news: true,
        model: currentModel,
      });

      // Parse metadata to separate text and chart data
      const cleanAnalysis = parseMetadata(data.analysis);
      setResult({ ...data, analysis: cleanAnalysis });

    } catch (err) {
      setError('获取预测失败，请检查股票代码或网络连接');
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // @ts-ignore
    if (route.params?.code) {
      // @ts-ignore
      const newCode = route.params.code;
      // @ts-ignore
      const initialModel = route.params.model || 'glm-4-flash-250414';
      // @ts-ignore
      const initialName = route.params.name || '';

      setCode(newCode);
      setModel(initialModel);
      setStockName(initialName);

      // Auto trigger prediction
      handlePredict(newCode, initialModel);
    }
  }, [route.params]);

  return (
    <View style={styles.container}>
      <Appbar.Header style={styles.header}>
        <Appbar.BackAction onPress={() => navigation.goBack()} color="#FFFFFF" />
        <Appbar.Content title="个股预测" titleStyle={styles.headerTitle} />
      </Appbar.Header>

      <ScrollView contentContainerStyle={styles.content}>
        <View style={styles.searchContainer}>
          <TextInput
            mode="outlined"
            label="股票代码 (如 sh600519)"
            value={code}
            onChangeText={setCode}
            style={styles.input}
            right={<TextInput.Icon icon="magnify" onPress={() => handlePredict(code, model)} />}
          />

          <Text style={styles.sectionTitle}>选择预测模型</Text>
          <SegmentedButtons
            value={model}
            onValueChange={setModel}
            buttons={[
              {
                value: 'glm-4-flash-250414',
                label: '深度分析',
                icon: 'brain',
              },
              {
                value: 'fractal',
                label: '分形预测',
                icon: 'chart-line-variant',
              },
            ]}
            style={styles.modelSelector}
          />

          <Button mode="contained" onPress={() => handlePredict(code, model)} loading={loading} style={styles.button}>
            开始预测
          </Button>
        </View>

        {error ? <HelperText type="error">{error}</HelperText> : null}

        {loading && <ActivityIndicator size="large" color="#1E88E5" style={{ marginTop: 20 }} />}

        {result && (
          <View>
            <Card style={styles.card}>
              <Card.Title title={model === 'fractal' ? "分形几何预测" : "AI 深度分析"} />
              <Card.Content>
                <Text variant="titleLarge" style={styles.stockTitle}>{stockName ? `${stockName} (${result.code})` : result.code}</Text>

                <View style={styles.confidenceContainer}>
                  <Text variant="bodyMedium">置信度: {(result.confidence * 100).toFixed(1)}%</Text>
                  <ProgressBar progress={result.confidence} color="#1E88E5" style={styles.progressBar} />
                </View>

                {model === 'fractal' && fractalData && (
                  <FractalChart data={fractalData} />
                )}

                {(() => {
                  const analysisSections = parseAnalysisSections(result.analysis);
                  return (
                    <>
                      <Text variant="titleMedium" style={styles.sectionTitle}>走势分析</Text>
                      <Text variant="bodyMedium" style={styles.analysisText}>{analysisSections.main}</Text>
                      {analysisSections.reasonLines.length > 0 ? (
                        <>
                          <Text variant="titleMedium" style={styles.sectionTitle}>智能体理由</Text>
                          {analysisSections.reasonLines.map((line, index) => (
                            <Text key={`${line}-${index}`} variant="bodySmall" style={styles.newsText}>{line}</Text>
                          ))}
                        </>
                      ) : null}
                    </>
                  );
                })()}

                {result.news_summary ? (
                  <>
                    {(() => {
                      const sections = parseSummarySections(result.news_summary);
                      return (
                        <>
                          {sections.main ? (
                            <>
                              <Text variant="titleMedium" style={styles.sectionTitle}>关键摘要</Text>
                              <Text variant="bodySmall" style={styles.newsText}>{sections.main}</Text>
                            </>
                          ) : null}
                          <Text variant="titleMedium" style={styles.sectionTitle}>冲突说明</Text>
                          <Text variant="bodySmall" style={styles.newsText}>{sections.conflict || '暂无冲突说明'}</Text>
                        </>
                      );
                    })()}
                  </>
                ) : null}
              </Card.Content>
            </Card>
          </View>
        )}
      </ScrollView>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    backgroundColor: '#1E88E5',
  },
  headerTitle: {
    color: '#FFFFFF',
    fontWeight: 'bold',
  },
  content: {
    padding: 16,
  },
  searchContainer: {
    marginBottom: 16,
  },
  input: {
    backgroundColor: '#FFFFFF',
    marginBottom: 12,
  },
  modelSelector: {
    marginBottom: 12,
  },
  button: {
    backgroundColor: '#1E88E5',
  },
  card: {
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    elevation: 2,
    marginBottom: 16,
  },
  stockTitle: {
    fontWeight: 'bold',
    marginBottom: 8,
    color: '#333',
  },
  confidenceContainer: {
    marginBottom: 16,
  },
  progressBar: {
    height: 8,
    borderRadius: 4,
    marginTop: 4,
  },
  sectionTitle: {
    fontWeight: 'bold',
    marginTop: 12,
    marginBottom: 4,
    color: '#1E88E5',
  },
  analysisText: {
    lineHeight: 22,
    color: '#424242',
  },
  newsText: {
    lineHeight: 20,
    color: '#616161',
  },
});

export default PredictionScreen;
