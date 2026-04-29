import { Radar, RadarChart as RechartsRadar, PolarGrid, PolarAngleAxis, PolarRadiusAxis, ResponsiveContainer } from 'recharts';
import type { ReportDimensions } from '../../types/report';

interface RadarChartProps {
  dimensions: ReportDimensions;
}

export const RadarChart = ({ dimensions }: RadarChartProps) => {
  const data = [
    { subject: '知识深度', value: dimensions.knowledge_depth, fullMark: 10 },
    { subject: '表达能力', value: dimensions.expression, fullMark: 10 },
    { subject: '问题解决', value: dimensions.problem_solving, fullMark: 10 },
    { subject: '代码质量', value: dimensions.code_quality, fullMark: 10 },
    { subject: '压力应对', value: dimensions.stress_response, fullMark: 10 },
  ];

  return (
    <ResponsiveContainer width="100%" height={400}>
      <RechartsRadar data={data}>
        <PolarGrid />
        <PolarAngleAxis dataKey="subject" />
        <PolarRadiusAxis angle={90} domain={[0, 10]} />
        <Radar
          name="能力评分"
          dataKey="value"
          stroke="#3b82f6"
          fill="#3b82f6"
          fillOpacity={0.6}
        />
      </RechartsRadar>
    </ResponsiveContainer>
  );
};
