package test

//
// Utilities For Table Based Testing Of The DataFlow Controller
//
// Test cases are a single instance of the controller's Reconcile().
// The controller-runtime "watch" functionality is not tested.
// The fake K8S client is used to track RuntimeObject state.
// The fake K8S client does NOT perform deletion of Owned RuntimeObjects.
//
/*
// TestCase Defines A Single Test Case (Initial/Final States)
type TestCase struct {
	// Descriptive Name For TestCase - Will Be Passed To t.Run()
	Name string

	// Flag Indicating That Only The Specified TestCase(s) Should Be Run
	// Meant to be bused as a temporary debug marker on a TestCase and not delivered!
	Only bool

	// The Namespace Of The RuntimeObject To Be Reconciled
	ReconcileNamespace string

	// The Name Of The RuntimeObject To Be Reconciled
	ReconcileName string

	// A Set Of RuntimeObjects To Setup In Cluster Prior To Reconciliation
	InitialState []runtime.Object

	// Optional Mock Implementation Of The Fake Client Get, List, Update, Delete
	//Mocks Mocks

	// The Expected Result (If Any) Returned From The Reconcile() Function
	ExpectedResult reconcile.Result

	// The Expected Error (If Any) Returned From The Reconcile() Function
	ExpectedError error

	// The Expected Events (If Any) Recorded During Reconciliation
	ExpectedEvents []corev1.Event

	// A Set Of RuntimeObjects Expected To Exist After Reconciliation
	ExpectedPresent []runtime.Object

	// A Set Of RuntimeObjects Expected NOT To Exit After Reconciliation
	ExpectedAbsent []runtime.Object

	// The Fake Client Reference
	client client.Client

	// The Fake EventRecorder Reference
	eventRecorder *MockEventRecorder
}

// Filter Specified TestCases To Only Those Marked With Only
func FilterTestCases(testCases []TestCase) []TestCase {
	onlyTestCases := make([]TestCase, 0)
	for _, testCase := range testCases {
		if testCase.Only {
			onlyTestCases = append(onlyTestCases, testCase)
		}
	}
	if len(onlyTestCases) == 0 {
		return testCases
	} else {
		return onlyTestCases
	}
}

// Get The TestCase's MockClient With Inner Fake Client With Initial State (Tests MUST Initialize scheme.Scheme!)
func (tc *TestCase) GetClient() client.Client {
	if tc.client == nil {
		tc.client = NewMockClient(fake.NewFakeClientWithScheme(scheme.Scheme, tc.InitialState...), tc.Mocks)
	}
	return tc.client
}

// Get The TestCase's MockEventRecorder
func (tc *TestCase) GetEventRecorder() *MockEventRecorder {
	if tc.eventRecorder == nil {
		tc.eventRecorder = NewMockEventRecorder()
	}
	return tc.eventRecorder
}

// Runner returns a testing func that can be passed to t.Run.
func (tc *TestCase) Runner(reconciler reconcile.Reconciler) func(t *testing.T) {

	// Return The Test Function For A Specific TestCase
	return func(t *testing.T) {

		// Perform The Reconciliation Of Desired RuntimeObject (Namespace / Name)
		result, err := reconciler.Reconcile(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: tc.ReconcileNamespace,
				Name:      tc.ReconcileName,
			},
		})

		// Clear The Mocks So That Verification Is Against "Real/Fake" K8S
		tc.client.(*MockClient).StopMocking()

		// Verify The Reconciliation Result / Error
		assert.Equal(t, tc.ExpectedResult, result)
		assert.Equal(t, tc.ExpectedError, err)

		// Verify The Expected Events Were Recorded
		tc.verifyExpectedEvents(t)

		// Verify The Expected Present & Absent RuntimeObjects
		tc.verifyExpectedPresent(t)
		tc.verifyExpectedAbsent(t)
	}
}

// Verify The TestCase's ExpectedEvent All Were Recorded
func (tc *TestCase) verifyExpectedEvents(t *testing.T) {

	// Loop Over All The ExpectedEvents - Checking Whether The Event Was Recorded
	for _, expectedEvent := range tc.ExpectedEvents {
		assert.True(t, tc.eventRecorder.EventRecorded(expectedEvent), "Expected Event Not Found: expectedEvent = %+v, events = %+v", expectedEvent, tc.eventRecorder.events)
	}
}

// Verify The TestCase's ExpectedPresent RuntimeObjects All Exist
func (tc *TestCase) verifyExpectedPresent(t *testing.T) {

	// Get The TestCase's Fake Client
	fakeClient := tc.GetClient()

	// Loop Over All The ExpectedPresent RuntimeObjects
	for _, expectedRuntimeObject := range tc.ExpectedPresent {

		// Create An Instance To Hold The Actual RuntimeObject Based On Expected RuntimeObject Type
		actualRuntimeObject, err := scheme.Scheme.New(expectedRuntimeObject.GetObjectKind().GroupVersionKind())
		assert.Nil(t, err)

		// Construct The ObjectKey
		accessor, err := meta.Accessor(expectedRuntimeObject)
		assert.Nil(t, err)
		objectKey := types.NamespacedName{
			Namespace: accessor.GetNamespace(),
			Name:      accessor.GetName(),
		}

		// Attempt To Get The RuntimeObject From The Fake Client
		err = fakeClient.Get(context.TODO(), objectKey, actualRuntimeObject)
		assert.Nil(t, err)

		// Create The Diff Comparison Options
		diffOpts := cmp.Options{

			// Ignore TypeMeta - Objects Created By Controller Should Have It's Not Required As K8S Will Populate
			cmpopts.IgnoreTypes(metav1.TypeMeta{}),

			// Ignore VolatileTime Fields - Will Have Millis Precision In Some Cases
			cmpopts.IgnoreTypes(apis.VolatileTime{}),

			// No Longer Necessary - Keeping In Case The Issue Reappears
			// cmpopts.IgnoreUnexported(resource.Quantity{}),
		}

		// Compare The Two RuntimeObjects
		diff := cmp.Diff(expectedRuntimeObject, actualRuntimeObject, diffOpts)
		if diff != "" {
			assert.Fail(t, fmt.Sprintf("unexpected present %T %s/%s (-want +got):\n%v", expectedRuntimeObject, accessor.GetNamespace(), accessor.GetName(), diff))
		}
	}
}

// Verify The TestCase's ExpectedAbsent RuntimeObjects Do NOT Exist
func (tc *TestCase) verifyExpectedAbsent(t *testing.T) {

	// Get The TestCase's Fake Client
	fakeClient := tc.GetClient()

	// Loop Over All The ExpectedAbsent RuntimeObjects
	for _, expectedRuntimeObject := range tc.ExpectedAbsent {

		// Create An Instance To Hold The Actual RuntimeObject Based On Expected RuntimeObject Type
		actualRuntimeObject, err := scheme.Scheme.New(expectedRuntimeObject.GetObjectKind().GroupVersionKind())
		assert.Nil(t, err)

		// Construct The ObjectKey
		accessor, err := meta.Accessor(expectedRuntimeObject)
		assert.Nil(t, err)
		objectKey := types.NamespacedName{
			Namespace: accessor.GetNamespace(),
			Name:      accessor.GetName(),
		}

		// Attempt To Get The RuntimeObject From The Fake Client & Verify Not Found Error
		err = fakeClient.Get(context.TODO(), objectKey, actualRuntimeObject)
		assert.NotNil(t, err)
		assert.True(t, apierrors.IsNotFound(err))
	}
}*/
