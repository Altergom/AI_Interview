import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../../components/layout/Container';
import { Button } from '../../components/common/Button';
import { Card } from '../../components/common/Card';
import { Loading } from '../../components/common/Loading';
import { RadarChart } from './RadarChart';
import { getReportStatus, getReport } from '../../services/report';
import { useInterviewStore } from '../../store/interviewStore';
import type { ReportResponse } from '../../types/report';

export const ReportView = () => {
  const navigate = useNavigate();
  const { interviewId } = useInterviewStore();

  const [report, setReport] = useState<ReportResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    if (interviewId) {
      pollReportStatus();
    }
  }, [interviewId]);

  const pollReportStatus = async () => {
    if (!interviewId) return;

    try {
      const status = await getReportStatus(interviewId);

      if (status.status === 'done') {
        const reportData = await getReport(interviewId);
        setReport(reportData);
        setLoading(false);
      } else if (status.status === 'failed') {
        setError('报告生成失败');
        setLoading(false);
      } else {
        setTimeout(pollReportStatus, 3000);
      }
    } catch (err) {
      setError('获取报告失败');
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <Container showHeader>
        <div className="max-w-4xl mx-auto py-16">
          <Loading size="lg" text="正在生成报告，请稍候..." fullScreen={false} />
        </div>
      </Container>
    );
  }

  if (error || !report) {
    return (
      <Container showHeader>
        <div className="max-w-4xl mx-auto py-16 text-center">
          <p className="text-red-600 mb-4">{error || '报告不存在'}</p>
          <Button onClick={() => navigate('/end')}>返回</Button>
        </div>
      </Container>
    );
  }

  return (
    <Container showHeader>
      <div className="max-w-4xl mx-auto py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-2">面试报告</h1>
        <p className="text-gray-600 mb-8">
          以下是您本次面试的综合评估
        </p>

        <div className="space-y-6">
          <Card>
            <h3 className="text-xl font-semibold mb-4">能力雷达图</h3>
            <RadarChart dimensions={report.dimensions} />
          </Card>

          <Card>
            <h3 className="text-xl font-semibold mb-4">总体评价</h3>
            <p className="text-gray-700">{report.summary}</p>
          </Card>

          <Card>
            <h3 className="text-xl font-semibold mb-4">优势点</h3>
            <ul className="space-y-2">
              {report.strong_points.map((point, index) => (
                <li key={index} className="flex items-start">
                  <span className="text-green-500 mr-2">✓</span>
                  <span className="text-gray-700">{point}</span>
                </li>
              ))}
            </ul>
          </Card>

          <Card>
            <h3 className="text-xl font-semibold mb-4">待改进点</h3>
            <ul className="space-y-2">
              {report.weak_points.map((point, index) => (
                <li key={index} className="flex items-start">
                  <span className="text-yellow-500 mr-2">!</span>
                  <span className="text-gray-700">{point}</span>
                </li>
              ))}
            </ul>
          </Card>
        </div>

        <div className="mt-8 flex justify-center">
          <Button onClick={() => navigate('/end')} size="lg">
            完成
          </Button>
        </div>
      </div>
    </Container>
  );
};
