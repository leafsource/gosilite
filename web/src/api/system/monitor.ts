import { request } from '@/utils/request';

import type {
  DatasourceDTO,
  FirewallResponse,
  HttpMonitorResponse,
  ServerDTO,
  SqlMonitorResponse,
  UriListResponse,
} from './monitorModel';

export function getDatasource() {
  return request.get<DatasourceDTO>({ url: '/system/monitor/datasource' });
}

export function getServer() {
  return request.get<ServerDTO>({ url: '/system/monitor/server' });
}

export function getSqlMonitor() {
  return request.get<SqlMonitorResponse>({ url: '/system/monitor/sql' });
}

export function getFirewall() {
  return request.get<FirewallResponse>({ url: '/system/monitor/firewall' });
}

export function getUriList() {
  return request.get<UriListResponse>({ url: '/system/monitor/uri' });
}

export function getHttpMonitor() {
  return request.get<HttpMonitorResponse>({ url: '/system/monitor/http' });
}
