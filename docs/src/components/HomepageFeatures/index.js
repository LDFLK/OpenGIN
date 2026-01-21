import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Temporal Data Model',
    image: require('@site/static/img/temporal_data_model.png').default,
    description: (
      <>
        Built-in support for time-travelling. Track both business time (validity)
        and system time (record time) to simulate digital twins with precision.
      </>
    ),
  },
  {
    title: 'Polyglot Storage',
    image: require('@site/static/img/polyglot_storage.png').default,
    description: (
      <>
        Automatically persists data in the most suitable format: Graph for relationships,
        Document for metadata, and Relational for tabular attributes.
      </>
    ),
  },
  {
    title: 'Specification',
    image: require('@site/static/img/specification.png').default,
    description: (
      <>
        Defines the blueprint for your entire ecosystem. Model organizations, businesses,
        and institutes with precision to create a comprehensive digital twin.
      </>
    ),
  },
];

function Feature({ image, title, description }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <img src={image} className={styles.featureSvg} alt={title} style={{ height: '200px', width: 'auto' }} />
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
