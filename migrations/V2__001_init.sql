INSERT INTO public.features (name) VALUES ('test-toggle');
INSERT INTO public.features (name) VALUES ('test-toggle-user-id');
INSERT INTO public.features (name) VALUES ('test-toggle-path');
INSERT INTO public.features (name) VALUES ('test-toggle-host');
INSERT INTO public.features (name) VALUES ('test-toggle-flexible-rollout-100');
INSERT INTO public.features (name) VALUES ('test-toggle-flexible-rollout-0');
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle', true);
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle-user-id', true);
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle-path', true);
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle-host', true);
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle-flexible-rollout-100', true);
INSERT INTO public.feature_environments (environment, feature_name, enabled) VALUES ('default', 'test-toggle-flexible-rollout-0', true);
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name) VALUES ('9f1c29d6-3b0d-4c4f-99d3-6bcd4d1d3f48', 'default', 'test-toggle', 'default', 'default');
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name, parameters) VALUES ('c3fbb5bc-75de-4d87-8c8e-7e2f403f4c3a', 'default', 'test-toggle-user-id', 'default', 'userWithId', '{"userIds": "12345"}');
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name) VALUES ('b1e389bb-6d93-4b51-b298-7080dea5cbac', 'default', 'test-toggle-path', 'default', 'default');
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name) VALUES ('4b0543de-9418-4edb-aa31-c649e705e752', 'default', 'test-toggle-host', 'default', 'default');
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name, parameters) VALUES ('7c36d122-59d2-4cc0-a8bb-f89300e661c5', 'default', 'test-toggle-flexible-rollout-100', 'default', 'flexibleRollout', '{"groupId": "test-toggle-flexible-rollout-100", "rollout": 100, "stickiness": "default"}');
INSERT INTO public.feature_strategies (id, project_name, feature_name, environment, strategy_name, parameters) VALUES ('0ac8e7aa-5d46-4b02-8e6c-1b122154c0a5', 'default', 'test-toggle-flexible-rollout-0', 'default', 'flexibleRollout', '{"groupId": "test-toggle-flexible-rollout-0", "rollout": 0, "stickiness": "default"}');