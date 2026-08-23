import type { AnalysisResult } from './api/types'

export type View = 'home' | 'inspect' | 'compare' | 'rehearse'

export interface RecentAnalysis {
  id: string
  kind: string
  at: string
  result: AnalysisResult
}
