/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.

 @type {import('@docusaurus/plugin-content-docs').SidebarsConfig}
 */
const sidebars = {
  // Manually defined sidebar to enforce order and naming
  docsSidebar: [
    {
      type: 'doc',
      id: 'index', // Home
      label: 'Home'
    },
    {
      type: 'category',
      label: 'Overview',
      collapsible: true,
      collapsed: false,
      items: [
        'overview/what_is_opengin',
        {
          type: 'doc',
          id: 'overview/why-opengin',
          label: 'Why OpenGIN?'
        },
      ],
    },
    {
      type: 'category',
      label: 'Architecture',
      collapsible: true,
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'overview/architecture/index',
          label: 'System Overview'
        },
        {
          type: 'doc',
          id: 'overview/architecture/data_flow',
          label: 'Data Flow'
        },
        {
          type: 'doc',
          id: 'overview/architecture/api-layer-details',
          label: 'Service APIs'
        },
        {
          type: 'doc',
          id: 'overview/architecture/core-api',
          label: 'Core API'
        },
        {
          type: 'doc',
          id: 'overview/architecture/database-schemas',
          label: 'Database Schemas'
        }
      ],
    },
    {
      type: 'category',
      label: 'Getting Started',
      collapsible: true,
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'getting_started/quick_start',
          label: 'Quick Start'
        },
        {
          type: 'doc',
          id: 'getting_started/installation',
          label: 'Installation'
        }
      ],
    },
    {
      type: 'category',
      label: 'Tutorial',
      collapsible: true,
      collapsed: true,
      items: [
        {
          type: 'doc',
          id: 'tutorial/simple_app',
          label: 'Simple Application'
        }
      ],
    },
    {
      type: 'category',
      label: 'Appendix',
      collapsible: true,
      collapsed: false,
      items: [
        {
          type: 'doc',
          id: 'appendix/datatype',
          label: 'Data Types'
        },
        {
          type: 'doc',
          id: 'appendix/data-type-detection-patterns',
          label: 'Type Detection'
        },
        {
          type: 'doc',
          id: 'appendix/storage',
          label: 'Storage Types'
        },
        {
          type: 'doc',
          id: 'appendix/limitations',
          label: 'Limitations'
        },
        {
          type: 'doc',
          id: 'appendix/release_life_cycle',
          label: 'Release Lifecycle'
        },
        {
          type: 'category',
          label: 'Operations',
          items: [
            {
              type: 'doc',
              id: 'appendix/operations/backup_integration',
              label: 'Backup Integration'
            },
            {
              type: 'doc',
              id: 'appendix/operations/mongodb',
              label: 'MongoDB Backup'
            },
            {
              type: 'doc',
              id: 'appendix/operations/neo4j',
              label: 'Neo4j Backup'
            },
            {
              type: 'doc',
              id: 'appendix/operations/postgres',
              label: 'PostgreSQL Backup'
            },
          ]
        }
      ],
    },
    {
      type: 'doc',
      id: 'faq',
      label: 'FAQ'
    }
  ],
};

export default sidebars;
