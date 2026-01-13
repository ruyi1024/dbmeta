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
    icon: 'DatabaseOutlined',
    path: '/meta/',
    component: './meta/',
  },
  {
    name: 'pumpkin',
    icon: 'BarChartOutlined',
    path: '/pumpkin',
    component: './pumpkin/index',
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
    name: 'aiAnalysis',
    icon: 'FileTextOutlined',
    path: '/task/analysis',
    component: './task/analysis',
  },
  {
    name: 'dataAlarm',
    icon: 'WarningOutlined',
    path: '/data/alarm',
    component: './data/alarm',
  },

  
  {
    name: 'aiDbQuery',
    icon: 'SearchOutlined',
    path: '/ai/dbquery',
    component: './ai/dbquery',
  },

  {
    name: 'aichat',
    icon: 'RobotOutlined',
    path: '/ai/chat',
    component: './ai/chat/',
  },
 
  // {
  //   name: 'event',
  //   icon: 'DashboardOutlined',
  //   path: '/event',
  //   component: './event/index',
  // },

  // {
  //   name: 'alarm',
  //   icon: 'WarningOutlined',
  //   path: '/alarm',
  //   //component: './Alarm/Index',
  //   routes: [
  //     { path: '/alarm', redirect: '/alarm/event' },
  //     {
  //       path: '/alarm/event',
  //       name: 'alarmEvent',
  //       component: './alarm/event',
  //       icon: 'MailOutlined',
  //     },
  //     {
  //       path: '/alarm/level',
  //       name: 'alarmLevel',
  //       component: './alarm/level',
  //       icon: 'OrderedListOutlined',
  //       //access: 'canAdmin',
  //     },
  //     {
  //       path: '/alarm/channel',
  //       name: 'alarmChannel',
  //       component: './alarm/channel',
  //       icon: 'UsergroupAddOutlined',
  //       //access: 'canAdmin',
  //     },
  //     {
  //       path: '/alarm/rule',
  //       name: 'alarmRule',
  //       component: './alarm/rule',
  //       icon: 'SettingOutlined',
  //       //access: 'canAdmin',
  //     },
  //     {
  //       path: '/alarm/suggest',
  //       name: 'alarmSuggest',
  //       component: './alarm/suggest',
  //       icon: 'HeartOutlined',
  //       //access: 'canAdmin',
  //     },
  //     {
  //       path: '/alarm/test',
  //       name: 'alarmTest',
  //       component: './alarm/test',
  //       icon: 'HeartOutlined',
  //       //access: 'canAdmin',
  //     },
  //     {
  //       path: '/alarm/nsq',
  //       name: 'nsqPage',
  //       component: './alarm/nsq',
  //       icon: 'HeartOutlined',
  //       //access: 'canAdmin',
  //     },
  //   ],
  // },
  {
    name: 'userManager',
    icon: 'UserOutlined',
    path: '/users/manager',
    component: './UserManager/index',
    access: 'canAdmin',
  },

  {
    name: 'task',
    icon: 'MenuUnfoldOutlined',
    path: '/task',
    component: './task/index',
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
      {
        path: '/setting/ai_models',
        name: 'ai_models',
        component: './ai/ai_models/',
        icon: 'RobotOutlined',
        access: 'canAdmin',
      },
      {
        path: '/setting/ai_chat_rules',
        name: 'ai_chat_rules',
        component: './ai/chat/rules/',
        icon: 'FileTextOutlined',
        access: 'canAdmin',
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
