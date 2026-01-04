export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        path: '/user',
        routes: [
          {
            name: 'login',
            path: '/user/login',
            component: './user/login',
          },
        ],
      },
    ],
  },
  { path: '/', redirect: '/portal/' },

  {
    name: 'portal',
    icon: 'ConsoleSqlOutlined',
    path: '/portal/',
    component: './portal/',
  },
  {
    name: 'execute',
    icon: 'ConsoleSqlOutlined',
    path: '/execute/',
    component: './execute/',
  },
  {
    name: 'meta',
    icon: 'database',
    path: '/meta',
    routes: [
      { path: '/meta', redirect: '/meta/dashboard' },
      {
        path: '/meta/dashboard',
        name: 'dashboard',
        component: './meta/dashboard',
        icon: 'BlockOutlined',
      },
      {
        path: '/meta/instance',
        name: 'instance',
        component: './meta/instance',
        icon: 'BlockOutlined',
      },
      {
        path: '/meta/database',
        name: 'database',
        component: './meta/database',
        icon: 'BlockOutlined',
      },
      {
        path: '/meta/table',
        name: 'table',
        component: './meta/table',
        icon: 'BlockOutlined',
      },
      {
        path: '/meta/column',
        name: 'column',
        component: './meta/column',
        icon: 'BlockOutlined',
      },
      {
        path: '/meta/quality',
        name: 'quality',
        component: './meta/quality',
        icon: 'BlockOutlined',
      },
    ],
  },

  {
    name: 'safe',
    icon: 'SafetyCertificateOutlined',
    path: '/safe',
    routes: [
      { path: '/safe', redirect: '/safe/dashboard' },
      {
        path: '/safe/dashboard',
        name: 'dashboard',
        component: './safe/dashboard',
        icon: 'BlockOutlined',
      },
      {
        name: 'audit',
        icon: 'AuditOutlined',
        path: '/safe/audit',
        component: './audit/',
      },
      {
        path: '/safe/privilege/grant/',
        name: 'privilege_grant',
        component: './privilege/grant/index',
        icon: 'CodeOutlined',
        access: 'canAdmin',
      },
      {
        path: '/safe/privilege/index/',
        name: 'privilege_index',
        component: './privilege/index',
        icon: 'SecurityScanOutlined',
      },
      {
        path: '/safe/sensitive/meta',
        name: 'sensitive_meta',
        component: './sensitive/meta',
        icon: 'SecurityScanOutlined',
      },
      {
        path: '/safe/sensitive/rule',
        name: 'sensitive_rule',
        component: './sensitive/rule',
        icon: 'OrderedListOutlined',
      },
    ],
  },

  {
    name: 'dataquality',
    icon: 'CheckCircleOutlined',
    path: '/dataquality',
    routes: [
      { path: '/dataquality', redirect: '/dataquality/dashboard' },
      {
        path: '/dataquality/dashboard',
        name: 'dashboard',
        component: './dataquality/dashboard',
        icon: 'DashboardOutlined',
      },
      {
        path: '/dataquality/issues',
        name: 'issues',
        component: './dataquality/issues',
        icon: 'ExclamationCircleOutlined',
      },
      {
        path: '/dataquality/rules',
        name: 'rules',
        component: './dataquality/rules',
        icon: 'SettingOutlined',
      },
      {
        path: '/dataquality/tasks',
        name: 'tasks',
        component: './dataquality/tasks',
        icon: 'PlayCircleOutlined',
      },
    ],
  },

  {
    name: 'aichat',
    icon: 'RobotOutlined',
    path: '/ai/chat',
    component: './ai/chat/',
  },
  {
    name: 'aichatrules',
    icon: 'RobotOutlined',
    path: '/ai/chat/rules',
    component: './ai/chat/rules/',
  },
  {
    name: 'aimodel',
    icon: 'RobotOutlined',
    path: '/ai/ai_models',
    component: './ai/ai_models/',
  },
  {
    name: 'aiAnalysis',
    icon: 'FileTextOutlined',
    path: '/task/analysis',
    component: './task/analysis',
  },
  {
    name: 'aiDbQuery',
    icon: 'SearchOutlined',
    path: '/ai/dbquery',
    component: './ai/dbquery',
  },


 
  {
    name: 'monitor',
    icon: 'DashboardOutlined',
    path: '/monitor',
    routes: [
      { path: '/monitor', redirect: '/monitor/dashboard' },
      {
        name: 'dashboard',
        icon: 'DashboardOutlined',
        path: '/monitor/dashboard/',
        component: './monitor/dashboard/index',
      },
      {
        name: 'event',
        icon: 'BarChartOutlined',
        path: '/monitor/event/',
        component: './monitor/event/index',
      },
      {
        name: 'processlist',
        icon: 'MonitorOutlined',
        path: '/monitor/processlist/',
        component: './monitor/processlist/simple',
      },
    ]
  },

  {
    name: 'alarm',
    icon: 'WarningOutlined',
    path: '/alarm',
    //component: './Alarm/Index',
    routes: [
      { path: '/alarm', redirect: '/alarm/event' },
      {
        path: '/alarm/event',
        name: 'alarmEvent',
        component: './alarm/event',
        icon: 'MailOutlined',
      },
      {
        path: '/alarm/level',
        name: 'alarmLevel',
        component: './alarm/level',
        icon: 'OrderedListOutlined',
        //access: 'canAdmin',
      },
      {
        path: '/alarm/channel',
        name: 'alarmChannel',
        component: './alarm/channel',
        icon: 'UsergroupAddOutlined',
        //access: 'canAdmin',
      },
      {
        path: '/alarm/rule',
        name: 'alarmRule',
        component: './alarm/rule',
        icon: 'SettingOutlined',
        //access: 'canAdmin',
      },
      {
        path: '/alarm/suggest',
        name: 'alarmSuggest',
        component: './alarm/suggest',
        icon: 'HeartOutlined',
        //access: 'canAdmin',
      },
      {
        path: '/alarm/test',
        name: 'alarmTest',
        component: './alarm/test',
        icon: 'HeartOutlined',
        //access: 'canAdmin',
      },
      {
        path: '/alarm/nsq',
        name: 'nsqPage',
        component: './alarm/nsq',
        icon: 'HeartOutlined',
        //access: 'canAdmin',
      },
    ],
  },
  {
    name: 'userManager',
    icon: 'UserOutlined',
    path: '/users/manager',
    component: './UserManager/index',
    access: 'canAdmin',
  },
  {
    name: 'setting',
    icon: 'SettingOutlined',
    path: '/setting',
    access: 'canAdmin',
    routes: [
      { path: '/setting', redirect: '/setting/datasource' },
      {
        path: '/setting/idc',
        name: 'idc',
        component: './setting/idc',
        icon: 'CloudServerOutlined',
        access: 'canAdmin',
      },
      {
        path: '/setting/env',
        name: 'env',
        component: './setting/env',
        icon: 'ChromeOutlined',
        access: 'canAdmin',
      },
            {
              path: '/setting/website',
              name: 'website',
              component: './setting/website',
              icon: 'GlobalOutlined',
              access: 'canAdmin',
            },
            {
              path: '/setting/api',
              name: 'api',
              component: './setting/api',
              icon: 'ApiOutlined',
              access: 'canAdmin',
            },
      {
        path: '/setting/datasource_type',
        name: 'datasource_type',
        component: './setting/datasource_type',
        icon: 'CodeSandboxOutlined',
        access: 'canAdmin',
      },
      {
        path: '/setting/datasource',
        name: 'datasource',
        component: './setting/datasource',
        icon: 'CloudOutlined',
        access: 'canAdmin',
      },

    ],
  },

  {
    name: 'task',
    icon: 'MenuUnfoldOutlined',
    path: '/task',
    routes: [
      { path: '/task', redirect: '/task/option' },
      {
        path: '/task/option',
        name: 'option',
        component: './task/option',
        icon: 'OrderedListOutlined',
      },
      {
        path: '/task/heartbeat',
        name: 'heartbeat',
        component: './task/heartbeat',
        icon: 'HeatMapOutlined',
      },
    ],
  },
  {
    name: 'change',
    icon: 'SwapOutlined',
    path: '/change',
    routes: [
      { path: '/change', redirect: '/change/query' },
      {
        path: '/change/query',
        name: 'query',
        component: './change/query',
        icon: 'SearchOutlined',
      },
    ],
  },
  {
    name: 'support',
    icon: 'BulbOutlined',
    path: '/support/',
    component: './support/index',
  },
  {
    component: './404',
  },
];
