export function generateSpecialistData() {
  const timestamp = Date.now();
  const randomId = Math.floor(Math.random() * 1000000);
  
  return {
    name: `Dr. Test User ${randomId}`,
    email: `test.user.${timestamp}.${randomId}@example.com`,
    phone: '+5511999999999',
    specialty: 'Cardiology',
    license_number: `CRM-SP-${timestamp}-${randomId}`,
    description: 'Experienced cardiologist for stress testing',
    keywords: ['cardiology', 'heart', 'stress-test'],
    agreed_to_share: true,
  };
}